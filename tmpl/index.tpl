<html>
<head>
    <link rel="stylesheet" href="/static/style.css" />
    <script src="//ajax.googleapis.com/ajax/libs/jquery/1.8.2/jquery.min.js"></script>
    <script src="/static/jquery-linedtextarea.js"></script>
    <title>Simpol Online Playground</title>
</head>
<body>
<div id="banner">
    <div id="head" itemprop="name">Simpol Online Playground</div>
    <div id="controls">
        <input type="button" value="Run" id="run">
        <input type="button" value="Share" id="share">
		<input type="checkbox" value="Debug" id="debug"> Debug
    </div>
</div>

<div style="position: absolute; top: 60px; font-size: 11px; width: 75%">
   A univesity project to create an interpreter for the <a href="https://github.com/edwindvinas/simpol">Simpol</a> language. This online interpreter allows users to test Simpol language online.
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
			var isDebug = document.getElementById('debug').checked;
            $.ajax({
                url: "/api/play?debug=" + isDebug,
                type: "POST",
                data: {
                    code: $('#code').val(),
                },
                success: function(data) {
                    $('#output').text(data)
                },
                error: function(data) {
                    $('#output').text(data.responseText)
                }
            })
        })
    })
-->
</script>
</body>
</html>
