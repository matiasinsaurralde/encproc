# ./build.sh: line 17: emcc: command not found
That error means emcc (the Emscripten compiler) isn’t in your shell’s PATH.

```console
git clone https://github.com/emscripten-core/emsdk.git
cd emsdk
./emsdk install latest
./emsdk activate latest
source ./emsdk_env.sh
```
 