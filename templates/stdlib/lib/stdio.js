/*
Contains JS implementations of <stdio.h> for both the server and the client
*/

{{define "stdio"}}
function putchar(c) {
  {{if .Server}}
  process.stdout.write(String.fromCharCode(c));
  {{else}}
  console.log(String.fromCharCode(c));
  {{end}}
}
{{end}}