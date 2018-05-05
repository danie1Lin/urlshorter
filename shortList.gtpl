<html>
<head>
    Shorter URL
</head>
<body>
<table width="300" border="1">
    <tr>
        <td>Short</td>
        <td>OriginalUrl</td>
        <td>Clicked</td>
    </tr>
    {{range .}}
    <tr>
        <td>{{.Short}}</td>
        <td><a href="{{.OriginalUrl}}">{{.OriginalUrl}}</a></td>
        <td>{{.Clicked}}</td>
    </tr>
    {{end}}
</table>
<form action="/short" method="post">
    <input type="text" placeholder="Enter your dick" name="url">
    <input type="submit" value="make it Shorter">
</form>
</body>
</html>
