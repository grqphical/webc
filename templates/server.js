{{define "server-js"}}
const fs = require("fs");
const bytes = fs.readFileSync("{{.BinaryName}}");

{{template "stdio" .}}

WebAssembly.instantiate(bytes, {
  libc: {
    putchar: putchar,
  },
}).then((results) => {
  const { main } = results.instance.exports;
  const exitCode = main();
  if (exitCode != 0) {
    console.error(
      `webc error: main() exited with non zero exit status: ${exitCode}`,
    );
  }
});
{{end}}