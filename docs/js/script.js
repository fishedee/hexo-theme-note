//挂载页面的事件
var ISMOBILE = false;

function binarySearchTocLower(list,elem){
  var left = 0;
  var right = list.length - 1;
  while( left <= right ){
    var middle = Math.floor((left+right)/2);
    var listMiddle = list[middle].target.offset().top;
    if( listMiddle == elem ){
      return middle;
    }else if( listMiddle > elem ){
      right = middle - 1;
    }else{
      left = middle + 1;
    }
  }
  if( left - 1 < 0 ){
    return 0;
  }else{
    return left - 1;
  }
}

function initToc(){
  var container  = $('#post');

  // Generate post TOC for h1 h2 and h3
  var toc = $('#post__toc-ul');
  var tocContainer = $('#post__toc');;
  var tocList = [];
  toc.empty();

  $('#post__content').find('h1,h2,h3').each(function() {
    var elem = $(this);
    var single = {};
    single.target = elem;
    single.tagName = {
      "h1":"h2",
      "h2":"h3",
      "h3":"h4",
    }[elem.prop("tagName").toLowerCase()];
    single.text = elem.text();
    single.id = elem.attr('id');
    tocList.push(single);
  });

  for( var i in tocList ){
    var singleToc = tocList[i];
    toc.append('<li class="post__toc-li post__toc-'+singleToc.tagName+'"><a href="#' + singleToc.id + '" class="js-anchor-link">' + singleToc.text + '</a></li>');
  }

  function updateToc(){
    var index = binarySearchTocLower(tocList,140);
    var targetToc = $('#post__toc .post__toc-li');
    targetToc.eq(index).addClass('active').siblings().removeClass('active');
    var targetScroll = targetToc.eq(index).offset().top - targetToc.eq(0).offset().top;
    tocContainer.scrollTop(targetScroll-tocContainer.height()/2);
  }
  container.off('scroll').bind('scroll',updateToc);
  updateToc();

  // Smooth scrolling
  $('.js-anchor-link').off('click').on('click', function() {
    var target = $(this.hash);
    container.animate({scrollTop: target.offset().top + container.scrollTop() - 140}, 500, function() {
      target.addClass('flash').delay(700).queue(function() {
        $(this).removeClass('flash').dequeue();
      });
    });
  });

  //toggle button
  $('#icon-list').off('click').on('click',function(){
    $('#post__toc').toggle('fast');
  });
  
}

function initHighLight(){
  var hightlight_interval = null;
  function checkHightlight() {
    if( !window.hljs ){
      return;
    }
    clearInterval(hightlight_interval);
    function getLinenumber(text){
      var linenumber = 0;
      var lastIndex = -1;
      for( var i = 0 ; i < text.length ; i++ ){
        var single = text.charAt(i);
        if( single == '\n'){
          ++linenumber;
          lastIndex = i;
        }
      }
      if( lastIndex != text.length -1 ){
        ++linenumber;
      }
      return linenumber;
    }
    function generateLineDiv(linenumber){
      var codeHtml = '<code style="float:left;" class="lineno">';
      for( var i = 1 ; i <= linenumber ; ++i ){
        codeHtml += i+"\n";
      }
      codeHtml += '</code>';
      return $(codeHtml);
    }
    function initLineNumber(){
      $('code').each(function(i, block) {
        block = $(block);
        if( block.hasClass("plantuml")){
          return;
        }
        if( block.hasClass("echarts")){
          return;
        }
        if( block.hasClass("flow")){
          return;
        }
        if( block.hasClass("sequence")){
          return;
        }
        if( block.hasClass("lineno") ){
          return;
        }
        if( block.prev().hasClass("lineno")){
          return;
        }
        if( block.parent().is('pre') == false ){
          console.log(block.parent().html());
          block.wrap('<pre></pre>');
          console.log(block.parent().html());
        }
        var linenumber = getLinenumber(block.text());
        var div = generateLineDiv(linenumber);
        block.before(div);
      });
    }
    function initHighLight(){
      hljs.configure({
        tabReplace: '    ', // 4 spaces
      });
      $('code').each(function(i, block) {
        hljs.highlightBlock(block);
      });
    }
    initLineNumber();
    initHighLight();
  }
  hightlight_interval = setInterval(checkHightlight,500);
}

function initMathJax(){
}

function initPlantUml(){
  $('code').each(function(i, block) {
    block = $(block);
    if( !block.parent().hasClass("plantuml")){
      return;
    }
    var ele = block.addClass('nohighlight').parent(); 
    ele.hide(); 
    var str = unescape(encodeURIComponent(block.text()));
    var imgURL = "http://www.plantuml.com/plantuml/svg/"+encode64(deflate(str,9));
    var newEle = $('<div><img src="'+imgURL+'" /></div>').insertAfter(ele);
  });
}

function initEcharts(){
  $('code').each(function(i, block) {
    block = $(block);
    if( !block.parent().hasClass("echarts")){
      return;
    }
    var ele = block.addClass('nohighlight').parent(); 
    ele.hide(); 
    var newEle = $('<div style="width: 100%;height:400px;"></div>');
    newEle.insertAfter(ele);  
    var myChart = echarts.init(newEle[0]);
    try  {
      eval(block.text());
      myChart.setOption(option);
    }catch(exception) {}
  });
}

function initFlow(){
  $('code').each(function(i, block) {
    block = $(block);
    if( !block.parent().hasClass("flow")){
      return;
    }
    var ele = block.addClass('nohighlight').parent(); 
    ele.hide(); 
    var newEle = $('<div style="width: 100%;"></div>');
    newEle.insertAfter(ele);
    var diagram = flowchart.parse(block.text());
    diagram.drawSVG(newEle[0]);
  });
}

function initSequence(){
  $('code').each(function(i, block) {
    block = $(block);
    if( !block.parent().hasClass("sequence")){
      return;
    }
    var ele = block.addClass('nohighlight').parent(); 
    ele.hide(); 
    var newEle = $('<div style="width: 100%;"></div>');
    newEle.insertAfter(ele);
    var diagram = Diagram.parse(block.text());
    diagram.drawSVG(newEle[0],{theme: 'simple'});
  });
}

function initSearch(){
  $('#search-form').submit(function(e){
    var target = $('#search-input');
    var text = target.val();
    var newText = text + ' site:fishedee.com';
    target.val(newText);
    setTimeout(function(){
      target.val(text);
    },500);
  });
}

initToc();
initSearch();
initHighLight();
initMathJax();
initPlantUml();
initEcharts();
initFlow();
initSequence();

