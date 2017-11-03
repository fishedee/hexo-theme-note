.PHONY:clean build server
build:clean
	cd blog/contents && make
	cd blog && hexo generate
	cp -r blog/public docs
	cp CNAME docs
clean:
	-rm -rf docs 
	-rm -rf blog/public/*
server:build
	cd blog && hexo server
