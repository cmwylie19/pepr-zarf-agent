Building

```bash
GOOS=js GOARCH=wasm go build -o main.wasm

cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" ../capabilities
```


**Important**

_modify wasm_exec.js and add polyfill for globalThis.crypto_

```js
// Initialize the polyfill
if (typeof globalThis.crypto === 'undefined') {
    globalThis.crypto = {
        getRandomValues: (array) => {
        for (let i = 0; i < array.length; i++) {
            array[i] = Math.floor(Math.random() * 256);
        }
        },
    };
}
```

Add "dom" to lib in `tsconfig.json`

```json
   "lib": [
      "ES2022","dom"
    ],
```
k create cm wasm --from-file=wasm.go=capabilities/main.wasm --dry-run=client -oyaml 


- How to mount main.wasm into the container
- How to pull funtions from wasm_exec.js into the container