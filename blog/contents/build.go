package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	. "github.com/fishedee/crypto"
	. "github.com/fishedee/language"
	. "github.com/fishedee/util"
	"io/ioutil"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"qiniupkg.com/api.v7/kodo"
	"strconv"
	"strings"
	"time"
)

var (
	isHelp          bool
	sourceDir       string
	targetDir       string
	blogAuthor      string
	blogTitle       string
	qiniuAccessKey  string
	qiniuSecretKey  string
	qiniuBucketName string
	qiniuDomain     string
)

func parseFlag() {
	flag.BoolVar(&isHelp, "help", false, "this help")
	flag.StringVar(&sourceDir, "input", "", "markdown source dirctory")
	flag.StringVar(&targetDir, "output", "", "html generate directory")
	flag.StringVar(&qiniuAccessKey, "qiniu_access", "", "qiniu access key")
	flag.StringVar(&qiniuSecretKey, "qiniu_secert", "", "qiniu secret key")
	flag.StringVar(&qiniuBucketName, "qiniu_bucket", "", "qiniu bucket name")
	flag.StringVar(&qiniuDomain, "qiniu_domain", "", "qiniu domain")
	flag.StringVar(&blogAuthor, "author", "", "blog article author")
	flag.StringVar(&blogTitle, "title", "", "blog title")
	flag.Parse()

	if isHelp ||
		sourceDir == "" ||
		targetDir == "" ||
		qiniuAccessKey == "" ||
		qiniuSecretKey == "" ||
		qiniuBucketName == "" ||
		qiniuDomain == "" ||
		blogAuthor == "" ||
		blogTitle == "" {
		fmt.Println(`version: build/1.0.0`)
		flag.Usage()
		os.Exit(0)
	}
}

type ProgressBar struct {
}

func (this *ProgressBar) Update(progress int, end int, desc string, args ...interface{}) {
	fmt.Printf("\r[%v/%v]%v", progress, end, desc)
}

func (this *ProgressBar) End() {
	fmt.Printf("\n")
}

type FileInfo struct {
	Content        string
	Path           string
	Date           string
	Name           string
	Title          string
	Category       string
	SecondCategory string
	Similar        float64
	IsDraft        bool
}

func isInt(in string) bool {
	_, err := strconv.Atoi(in)
	return err == nil
}

func runCmd(stdin string, name string, args ...string) ([]byte, error) {
	var buf = bytes.NewBuffer([]byte(""))
	cmd := exec.Command(name)
	cmd.Stdin = strings.NewReader(stdin)
	cmd.Stdout = buf
	cmd.Stderr = os.Stderr
	cmd.Args = append([]string{name}, args...)
	cmd.Env = os.Environ()
	err := cmd.Run()
	return buf.Bytes(), err
}

func convertHtml(data string) string {
	resultByte, err := runCmd(data, "pandoc", "-f", "markdown", "-t", "html", "--mathjax")
	if err != nil {
		panic(err)
	}
	result := string(resultByte)
	return result
}

func convertSpecial(result string) string {
	result = strings.Replace(result, "{{", "&#123&#123", -1)
	result = strings.Replace(result, "}}", "&#125&#125", -1)
	result = strings.Replace(result, `<pre class=" ">`, "<pre>", -1)
	return result
}

func getDirFiles(dir string) []string {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		panic(err)
	}
	result := []string{}
	for _, file := range files {
		if file.Name() == ".DS_Store" {
			continue
		}
		if file.IsDir() == true {
			subDir := dir + "/" + file.Name()
			subFile := getDirFiles(subDir)
			result = append(result, subFile...)
		} else {
			fileName := dir + "/" + file.Name()
			result = append(result, fileName)
		}
	}
	return result
}

func readDir(dir string) []FileInfo {
	files := getDirFiles(dir)
	result := []FileInfo{}
	progress := &ProgressBar{}
	for index, file := range files {
		progress.Update(index+1, len(files), "Reading file ...", file)
		defer CatchCrash(func(e Exception) {
			fmt.Printf("\nfile:%v,readFile error:%v", file, e)
		})
		FileInfo := readFile(file)
		if FileInfo.IsDraft {
			continue
		}
		result = append(result, FileInfo)
	}
	progress.End()
	return result
}

func readFile(path string) FileInfo {
	//解析path
	result := FileInfo{}
	result.Path = path
	basePath := filepath.Base(path)
	baseExt := filepath.Ext(basePath)
	basePath = basePath[:len(basePath)-len(baseExt)]
	filePathSeg := Explode(basePath, "-")
	if len(filePathSeg) < 4 ||
		isInt(filePathSeg[0]) == false ||
		isInt(filePathSeg[1]) == false ||
		isInt(filePathSeg[2]) == false {
		Throw(1, "Invalid Path "+path)
	}
	result.Date = filePathSeg[0] + "-" + filePathSeg[1] + "-" + filePathSeg[2]
	result.Title = Implode(filePathSeg[3:], "-")
	result.Name = strings.Replace(result.Title, " ", "-", -1)
	if strings.HasPrefix(result.Name, "draft_") {
		result.IsDraft = true
	} else {
		result.IsDraft = false
	}
	dirPath := Explode(filepath.Dir(result.Path), "/")
	if len(dirPath) < 2 {
		result.Category = ""
	} else {
		result.Category = dirPath[1]
	}
	if len(dirPath) < 3 {
		result.SecondCategory = ""
	} else {
		result.SecondCategory = dirPath[2]
	}

	//解析内容
	dateByte, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	result.Content = string(dateByte)
	return result
}

func min(args ...int) int {
	result := args[0]
	for i := 1; i < len(args); i++ {
		if args[i] < result {
			result = args[i]
		}
	}
	return result
}

func max(args ...int) int {
	result := args[0]
	for i := 1; i < len(args); i++ {
		if args[i] > result {
			result = args[i]
		}
	}
	return result
}

func abs(a int) int {
	if a < 0 {
		return -a
	} else {
		return a
	}
}

func calcCategorySimilarity(category11 string, category12 string, category21 string, category22 string) float64 {
	if category11 == category21 {
		if category12 == category22 {
			return 1
		} else {
			return 0.8
		}
	}
	return 0
}

func calcDateSimilarity(a string, b string) float64 {
	lhs, err := time.Parse("2006-01-02", a)
	if err != nil {
		panic(err)
	}
	rhs, err := time.Parse("2006-01-02", b)
	if err != nil {
		panic(err)
	}
	diff := lhs.Sub(rhs) / (time.Hour * 24)
	return math.Pow(float64(abs(int(diff+1))), -0.2)
}

func calcStringSimilarity(a string, b string) float64 {
	lhs := []rune(a)
	rhs := []rune(b)
	dp := make([][]int, len(lhs)+1, len(lhs)+1)
	for i := 0; i <= len(lhs); i++ {
		dp[i] = make([]int, len(rhs)+1, len(rhs)+1)
		dp[i][0] = i
	}
	for j := 0; j <= len(rhs); j++ {
		dp[0][j] = j
	}
	for i := 1; i <= len(lhs); i++ {
		for j := 1; j <= len(rhs); j++ {
			cur := 0
			if lhs[i-1] != rhs[j-1] {
				cur = 1
			} else {
				cur = 0
			}
			dp[i][j] = min(dp[i-1][j-1]+cur, dp[i-1][j]+1, dp[i][j-1]+1)
		}
	}
	diff := dp[len(lhs)][len(rhs)]
	if diff == 0 {
		return 1
	}
	return math.Pow(float64(diff), -2.0)
}

func getSimilarPost(target FileInfo, allFileInfo []FileInfo) []FileInfo {
	for index, singleFileInfo := range allFileInfo {
		if singleFileInfo.Path == target.Path {
			allFileInfo[index].Similar = 0.0
			continue
		}
		nameSim := calcStringSimilarity(target.Name, singleFileInfo.Name)
		dateSim := calcDateSimilarity(target.Date, singleFileInfo.Date)
		categorySim := calcCategorySimilarity(
			target.Category,
			target.SecondCategory,
			singleFileInfo.Category,
			singleFileInfo.SecondCategory,
		)
		allFileInfo[index].Similar = 30*categorySim + 10*nameSim + dateSim
	}
	return QuerySort(allFileInfo, "Similar Desc").([]FileInfo)
}

type HandleFileInfo struct {
	Content        string
	Date           string
	Category       string
	SecondCategory string
	Name           string
	Title          string
}

func handleAllCategory(indexFile HandleFileInfo, data []HandleFileInfo) string {
	data = QuerySort(data, "Date Desc").([]HandleFileInfo)
	content := indexFile.Content
	lastDate := ""
	lastDateIndex := 0
	content += "\n# 分类\n\n"
	QueryGroup(data, "Category Desc", func(singleCategory []HandleFileInfo) int {
		content += fmt.Sprintf("* [%v](/categories/%v)\n", singleCategory[0].Category, singleCategory[0].Category)
		return 1
	})
	for _, singleData := range data {
		dateInfo := Explode(singleData.Date, "-")
		month, _ := strconv.Atoi(dateInfo[1])
		dateDesc := ""
		if month <= 6 {
			dateDesc = dateInfo[0] + "上半年"
		} else {
			dateDesc = dateInfo[0] + "下半年"
		}
		if lastDateIndex == 0 ||
			lastDate != dateDesc {
			lastDateIndex++
			lastDate = dateDesc
			content += fmt.Sprintf("\n# %v \n\n", lastDate)
		}
		content += fmt.Sprintf("* [%v-%v](/%v/%v)\n", singleData.Date, singleData.Title, Implode(Explode(singleData.Date, "-"), "/"), singleData.Name)
	}
	yaml := "---\n" +
		"layout: post\n" +
		"title: " + blogTitle + "\n" +
		"date: 2017-11-01\n" +
		"---\n\n"
	return yaml + convertHtml(content)
}

func HandleSingleCategory(data []HandleFileInfo) string {
	data = QuerySort(data, "SecondCategory Desc,Date Desc").([]HandleFileInfo)
	content := ""
	lastCategory := ""
	lastCategoryIndex := 0
	for _, singleData := range data {
		if lastCategoryIndex == 0 ||
			lastCategory != singleData.SecondCategory {
			lastCategoryIndex++
			lastCategory = singleData.SecondCategory
			categoryName := ""
			if lastCategory == "" {
				categoryName = "其他"
			} else {
				categoryName = lastCategory
			}
			content += fmt.Sprintf("\n# %v \n\n", categoryName)
		}
		content += fmt.Sprintf("* [%v-%v](/%v/%v)\n", singleData.Date, singleData.Title, Implode(Explode(singleData.Date, "-"), "/"), singleData.Name)
	}
	yaml := "---\n" +
		"layout: post\n" +
		"title: " + data[0].Category + "\n" +
		"date: 2017-11-01\n" +
		"---\n\n"
	return yaml + convertHtml(content)
}

func handleFile(singleFile FileInfo, allFile []FileInfo) (*HandleFileInfo, string) {
	result := &HandleFileInfo{}
	result.Category = singleFile.Category
	result.SecondCategory = singleFile.SecondCategory
	result.Name = singleFile.Name
	result.Title = singleFile.Title
	result.Date = singleFile.Date
	result.Content = singleFile.Content

	yaml := "---\n" +
		"layout: post\n" +
		"category: " + result.Category + "\n" +
		"title: " + result.Title + "\n" +
		"date: " + result.Date + "\n" +
		"---\n\n"

	content := singleFile.Content
	content += "\n\n" +
		"> * 本文作者： " + blogAuthor + "\n" +
		"> * 版权声明： 本博客所有文章均采用 [CC BY-NC-SA 3.0 CN 许可协议](https://creativecommons.org/licenses/by-nc-nd/3.0/deed.zh)，转载必须注明出处！\n"
	content += "\n# 相关文章\n\n"
	for index, similarFile := range getSimilarPost(singleFile, allFile) {
		content += fmt.Sprintf("* [%v-%v](/%v/%v)\n", similarFile.Date, similarFile.Title, Implode(Explode(similarFile.Date, "-"), "/"), similarFile.Name)
		if index >= 4 {
			break
		}
	}

	return result, yaml + convertHtml(content)
}

func handleDir(files []FileInfo) map[string]string {
	handleFiles := []HandleFileInfo{}
	indexFile := HandleFileInfo{}
	result := map[string]string{}
	progress := &ProgressBar{}
	for index, singleFile := range files {
		progress.Update(index+1, len(files), "Handle File ...")
		defer CatchCrash(func(e Exception) {
			fmt.Printf("\nfile:%v,handleFile error:%v\n", singleFile.Name, e)
		})
		singleHandleFile, data := handleFile(singleFile, files)
		if singleHandleFile.Category != "" {
			handleFiles = append(handleFiles, *singleHandleFile)
			fileName := "_posts/" + singleHandleFile.Name + ".html"
			_, isExist := result[fileName]
			if isExist == true {
				Throw(1, "Has Exist File!"+fileName)
			}
			result[fileName] = data
		} else if singleHandleFile.Name == "index" {
			indexFile = *singleHandleFile
		} else {
			result[singleHandleFile.Name+".html"] = data
		}
	}
	progress.End()

	allCategoryData := handleAllCategory(indexFile, handleFiles)
	result["index.html"] = allCategoryData

	QueryGroup(handleFiles, "Category desc", func(data []HandleFileInfo) int {
		categoryName := data[0].Category
		categoryData := HandleSingleCategory(data)
		result["categories/"+categoryName+"/index.html"] = categoryData
		return 1
	})
	return result

}

type imageInfo struct {
	origin string
	target string
	file   string
}

func uploadSingleImage(bucket kodo.Bucket, image *imageInfo) {
	defer CatchCrash(func(e Exception) {
		fmt.Printf("\nfile:%v,image: %v,error:%v\n", image.file, image.origin, e)
	})
	src := image.origin
	if len(src) == 0 {
		panic("image is empty")
	}
	//非本地文件
	if strings.HasPrefix(src, "http") {
		image.target = image.origin
		return
	}
	//文件已经存在
	srcSha := CryptoSha1([]byte(src))
	_, err := bucket.Stat(nil, srcSha)
	if err == nil {
		image.target = "http://" + qiniuDomain + "/" + srcSha
		return
	}
	src = src[1:]
	src = strings.Replace(src, "%20", " ", -1)
	//文件不存在，上传
	imageData, err := ioutil.ReadFile(src)
	if err != nil {
		panic(err)
	}
	putRet := kodo.PutRet{}
	err = bucket.Put(nil, &putRet, srcSha, bytes.NewReader(imageData), int64(len(imageData)), &kodo.PutExtra{})
	if err != nil {
		panic(err)
	}
	image.target = "http://" + qiniuDomain + "/" + srcSha
	return
}

func uploadImage(files map[string]string) map[string]string {
	docFiles := map[string]*goquery.Document{}
	images := map[string]*imageInfo{}
	for singleFileName, singleFile := range files {
		doc, err := goquery.NewDocumentFromReader(bytes.NewBufferString(singleFile))
		if err != nil {
			panic(err)
		}
		doc.Find("img").Each(func(i int, s *goquery.Selection) {
			src := s.AttrOr("src", "")
			images[src] = &imageInfo{
				file:   singleFileName,
				origin: src,
				target: "",
			}
		})
		docFiles[singleFileName] = doc
	}

	cfg := &kodo.Config{
		AccessKey: qiniuAccessKey,
		SecretKey: qiniuSecretKey,
	}
	qiniuClient := kodo.New(0, cfg)
	qiniuBucket := qiniuClient.Bucket(qiniuBucketName)

	progress := &ProgressBar{}
	task := NewTask()
	task.SetIsAutoStop(true)
	task.SetThreadCount(16)
	task.SetBufferCount(1024)
	task.SetProgressHandler(func(now int, max int) {
		progress.Update(now, max, "Upload image ...")
	})
	task.SetHandler(uploadSingleImage)
	task.Start()
	for _, singleImageInfo := range images {
		task.AddTask(qiniuBucket, singleImageInfo)
	}
	task.Wait()
	progress.End()

	result := map[string]string{}
	for singleFileName, singleDocFile := range docFiles {
		singleDocFile.Find("img").Each(func(i int, s *goquery.Selection) {
			src := s.AttrOr("src", "")
			targetSrc := images[src].target
			s.SetAttr("src", targetSrc)
		})
		html, err := singleDocFile.Find("body").Html()
		if err != nil {
			panic(err)
		}
		result[singleFileName] = convertSpecial(html)
	}
	return result
}

func writeDir(newDir string, data map[string]string) {
	err := os.RemoveAll(newDir)
	if err != nil {
		panic(err)
	}

	for file, content := range data {
		dir := filepath.Dir(file)
		err := os.MkdirAll(newDir+"/"+dir, os.ModePerm)
		if err != nil {
			panic(err)
		}
		err = ioutil.WriteFile(newDir+"/"+file, []byte(content), os.ModePerm)
		if err != nil {
			panic(err)
		}
	}
}

func main() {
	parseFlag()
	defer CatchCrash(func(e Exception) {
		fmt.Println(e)
	})

	inputFiles := readDir(sourceDir)
	outputFiles := handleDir(inputFiles)
	outputFiles = uploadImage(outputFiles)
	writeDir(targetDir, outputFiles)
}
