{{define "server-js"}}
const fs = require("fs");
const bytes = fs.readFileSync("{{.BinaryName}}");

{{template "stdio" .}}
{{template "math" .}}

WebAssembly.instantiate(bytes, {
  libc: {
    putchar: putchar,

    fabsf: fabsf,
    fmodf: fmodf,
    remainderf: remainderf,
    expf: expf,
    exp2f: exp2f,
    expm1f: expm1f,
    logf: logf,
    log10f: log10f,
    log2f: log2f,
    log1pf: log1pf,
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