/*
Contains JS implementations of <stdio.h> for both the server and the client
*/

{{define "stdio"}}
function putchar(c) {
  process.stdout.write(String.fromCharCode(c));
}
{{end}}