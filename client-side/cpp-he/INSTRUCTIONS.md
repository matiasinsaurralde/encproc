# ./buildSmall

Does not work yet. The idea is to get smaller wasm binaries...It produces only marginally smaller binaries. (1.5MB --> 1.2 MB)

# ./build.sh: line 17: emcc: command not found
That error means emcc (the Emscripten compiler) isn’t in your shell’s PATH.

```console
git clone https://github.com/emscripten-core/emsdk.git
cd emsdk
./emsdk install latest
./emsdk activate latest
source ./emsdk_env.sh
```
 
# wasm-ld: error: unable to find libraries .a
If you just ran a native make install into /usr/local, you’ll have a native x86_64 .a/.so in /usr/local/lib. That won’t work. You need to install OpenFHE in a suitable way.

Clone OpenFHE somewhere (e.g. ~/openfhe-wasm-build) and build it with Emscripten:
```console
cd ~/openfhe-wasm-build
mkdir emscripten_build && cd emscripten_build
emcmake cmake .. -DCMAKE_BUILD_TYPE=Release -DCMAKE_INSTALL_PREFIX=$(pwd)/install
emmake make -j
emmake make install
```
After that you should have
```console
~/openfhe-wasm-build/emscripten_build/install/lib/libopenfhe.a
~/openfhe-wasm-build/emscripten_build/install/include/openfhe/...
```

At last
```console
export OPENFHE_ROOT=~/openfhe-wasm-build/emscripten_build/install
```
OR
```console
OPENFHE_ROOT=/path/to/emscripten_build/install ./build.sh
```

(You can link the .a file directly otherwise)