package nopaste

import "html/template"

var tmpl = template.Must(template.New("tmpl").Parse(`{{define "index"}}
<!DOCTYPE html>
<html>
  <head>
    <meta charset="utf-8">
    <title>Nopaste</title>
    <style type="text/css">
<!--
body {
    font-family: Verdana, sans-serif;
    font-size: 0.9em;
    background-color: #fff;
    padding: 20px;
}

h1 {
    font-family: "Trebuchet MS", Helvetica, Arial, sans-serif;
    font-weight: bold;
    font-size: 1.6em;
    padding-bottom: 20px;
}

form p {
    padding: 5px 0;
}

label {
    padding-right: 5px;
}

input, select {
    margin-right: 20px;
}
-->
      </style>
    </head>
  <body>
  <h1>Nopaste</h1>
    <form method="post">
    <p>
       <label for="channel">channel:</label>
       <select id="channel" name="channel">
         <option>(None)</option>
       {{range .Channels}}
         <option value="{{.}}">{{.}}</option>
       {{end}}
       </select>
       <label for="nick">nickname:</label><input type="text" name="nick">
    </p>
    <p>
       <label for="summary">summary:</label>
       <input type="text" name="summary" size="80">
    </p>
    <p>
       <textarea id="text" name="text" rows="24" cols="100"></textarea>
    </p>
    <p>
       <label for="notice">notice</label>
       <input type="checkbox" name="notice" value="1">
    </p>
    <p>
       <input type="submit" value="Paste it">
    </p>
    </form>
  </body>
</html>
{{end}}
`))
