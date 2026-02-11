{{define "browser-js"}}

{{template "stdio" .}}

WebAssembly.instantiateStreaming(fetch("/{{.BinaryName}}"), {
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