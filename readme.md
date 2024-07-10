# CBuild
A C build system ~~that's definitely not discount CMake~~ written in Go

I had *way* too much fun writing this

## Installation
**Requirements**
* Go >= 1.22.1

The only way to Install CBuild currently is to build it from source:

First, install Go if you haven't already from [go.dev](go.dev)

Then, run this command:
```sh
git clone "https://github.com/voidwyrm-2/CBuild" && cd ./CBuild && sh ./build.sh
```

## Usage
If you want to run the test I have in this repo, do(assuming you just followed the installation instructions):
```sh
cd test && cbuild
```
Refer to the [CBuildfile example](./CBuildfile_example.txt) for instructions how to use CBuildfiles<br>
Use `cbuild -h` for command help

**If you have an issue, report it in issues!**

### This repo is licensed under the [MIT license](./LICENSE)