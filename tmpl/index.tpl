<html>
<head>
    <link rel="stylesheet" href="/static/style.css" />
    <script src="//ajax.googleapis.com/ajax/libs/jquery/1.8.2/jquery.min.js"></script>
    <script src="/static/jquery-linedtextarea.js"></script>
    <title>Simpol Interpreter</title>
</head>
<body>
<div id="banner">
    <div id="head" itemprop="name">Simpol Interpreter</div>
    <div id="controls">
		<input type="button" value="Home" id="home" title="Go to home">
        <input type="button" value="Run" id="run" title="Execute the code"> 
        <input type="button" value="Share" id="share" title="Share the code">
		<input type="button" value="Explore" id="explore" title="List sample codes">
		<input type="checkbox" value="Debug" id="debug" title="Show debug information"> Debug
    </div>
</div>

<div style="position: absolute; top: 60px; font-size: 11px; width: 75%">
   A univesity project to create an interpreter for the <a href="https://github.com/edwindvinas/simpol">Simpol</a> language. This online interpreter allows users to test Simpol language online. It uses Google Appengine, Golang, and hand-written scanning and parsing logic. Please note that this is not the ideal solution for interpreters which usually use tools such as EBNF, Lex, Yacc, etc in order to create a robust interpreter software. But using these tools are complex to implement in a university project.
</div>

<div id="wrap">
    <textarea itemprop="description" id="code" name="code" autocorrect="off" autocomplete="off" autocapitalize="off" spellcheck="false">{{.Code}}</textarea>
</div>

<div><pre id="output"></pre></div>

<script type="text/javascript">
<!--
    $(document).ready(function() {
        $('#code').linedtextarea();
        $('#share').click(function() {
            $.ajax({
                url: "/api/save",
                type: "POST",
                data: {
                    code: $('#code').val(),
                },
                success: function(data) {
                    location.href = "/p/" + data
                },
                error: function(data) {
                    alert(data.responseText)
                }
            })
        })
        $('#run').click(function() {
			//if user input is needed; ask them before submission
			var codeInput = $('#code').val();
			console.log(codeInput);
			var res = codeInput.split("\n");
			var ioj = "";
			for(var i=0; i < res.length; i++){
				console.log(res[i]);
				//if ASK input
				var str = res[i];
				var n = str.indexOf("ASK");
				if (n >= 0) {
					var sim = str.split("ASK");
					if (sim[1] != "") {
						aid = sim[1];
						console.log(aid);
						thisIn = prompt("Enter ASK value", aid.trim());
						ioj = ioj + "|" + aid + "=" + thisIn;
					}
				}

			}
			console.log(ioj);
			var isDebug = document.getElementById('debug').checked;
            $.ajax({
                url: "/api/play?debug=" + isDebug,
                type: "POST",
                data: {
                    code: $('#code').val(),
					input: ioj,
                },
                success: function(data) {
                    $('#output').text(data)
                },
                error: function(data) {
                    $('#output').text(data.responseText)
                }
            })
        })
        $('#home').click(function() {
			location.href = "http://simpol-online.appspot.com";
        })
        $('#explore').click(function() {
			location.href = "http://simpol-online.appspot.com/explore";
        })
    })
-->
</script>
</body>
</html>
