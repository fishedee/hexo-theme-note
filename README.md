# hexo-theme-note

学习笔记式的hexo主题，浏览[这里](http://fishedee.com)

# 功能

* 数学公式与md文件混合排版
* 自动生成每篇文章的相关推荐
* 自动生成每篇文章的版权声明
* 自动生成分类和时间信息，不再需要填写麻烦的yaml头
* 自动生成首页和分类页面
* 自动将图片上传到七牛图床，并相应更新markdown中的图片链接
* 标志为草稿箱文章，不发布

# 使用

* 安装node，hexo-cli，以及cd blog&& npm install
* 安装golang和[fishgo](https://github.com/fishedee/fishgo)
* 安装pandoc
* 在blog/contents/_post下写入你的markdown文件
* 在blog/contents/Makefile配置你的七牛图床key与secert
* 在根目录make即可，生成文件放到了docs目录下

# 原理

```
md -> html
```

原来hexo的使用方式是将source下的文件编译为html文件，但这样做很麻烦，md文件需要逐个写自己的yaml头，而且默认的渲染引擎遇到复杂数学公式排版就会崩。

```
md -> html -> html
```

所以改为一个go脚本，将contents下的md文件，转换为source下的带yaml头的html文件，然后由hexo进一步转换为最终的html文件。

其实，脚本还能一步到位直接生成到最终的html文件，不过这样就利用不了hexo的主题库了，所以还是由hexo来生成最终的html文件好了。当然，这样的方法也能套到jekyll上，只要将脚本单独拉出来用就可以了。
