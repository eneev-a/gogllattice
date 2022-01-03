# Go GL lattice

gogllattice is a learning project for OpenGL

lattice.go generates a color cube lattice where each individual cube's
size is cycled between small and large producing fractal-like images.

Controls are `W`, `A`, `S`, `D`. `Space` for "up", and `Z` for "down".
`Shift`+key reduces speed. `Ctrl`+key increases speed.

## To run on Linux:

```sh
go run ./lattice.go
```

# Cross-compile for Windows

```sh
CC=x86_64-w64-mingw32-gcc CXX=x86_64-w64-mingw32-g++ GOOS=windows CGO_ENABLED=1 go build ./lattice.go
```
