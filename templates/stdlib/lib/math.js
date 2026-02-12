{{define "math"}}
function fabsf(arg) {
    return Math.abs(arg)
}

function fmodf(x, y) {
    return x % y
}

function remainderf(x, y) {
    return x % y
}

function expf(x) {
    return Math.pow(Math.E, x)
}

function exp2f(x) {
    return Math.pow(2, x)
}

function expm1f(x) {
    return Math.pow(Math.E, x) - 1;
}

function logf(x) {
    return Math.log(x)
}

function log10f(x) {
    return Math.log10(x)
}

function log2f(x) {
    return Math.log2(x)
}

function log1pf(x) {
    return Math.log(1+x)
}

{{end}}