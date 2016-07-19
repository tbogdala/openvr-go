Openvr-go
=========

Openvr-go is an [Go][golang] programming language wrapper for the [OpenVR SDK][openvr-git]
published by Valve for VR hardware.


UNDER CONSTRUCTION
==================

At present, it is very much in an alpha stage with new development happening to
complete the API exposed by the OpenVR SDK.


Requirements
------------

The wrapper library itself doesn't have any dependencies. The `connectiontest` sample
in the `examples` folder also doesn't have any dependencies.

The other samples are graphical and use the following libraries, though they are
not imported by the core openvr-go module itself:

* [Mathgl][mgl] - for 3d math
* [GLFW][glfw-go] (v3.1) - creating windows and providing the OpenGL context
* [Fizzle][fizzle] - provides the graphics engine
* [Go GL][go-gl] - provides the backend implementation of OpenGL for [Fizzle][fizzle].

Installation
------------

The dependency Go libraries for graphical examples can be installed with the following commands.

```bash
go get github.com/go-gl/glfw/v3.1/glfw
go get github.com/go-gl/mathgl/mgl32
go get github.com/go-gl/gl/v3.3-core/gl
go get github.com/tbogdala/fizzle
```
This does assume that you have the native GLFW 3.1 library installed already
accessible to Go tools.

Additionally, the appropriate `openvr_api.dll` or `libopenvr_api.so` file from
`vendor/openvr/bin/<platform>` will either need to be copied into each example directory
being built or it will need to be accessible system wide.

Each sample can be built by going to that directory in a shell and executing
a `go build` command. For example:

```bash
cd $GOPATH/src/github.com/tbogdala/openvr-go/examples/basiccube
go build
cp ../../vendor/openvr/bin/win64/openvr_api.dll .
./basiccube.exe
```

Current Features
----------------

Partial implementation of the following interfaces:

* IVRSystem
* IVRCompositor
* IVRRenderModels


LICENSE
=======

Original source code in openvr-go is released under the BSD license. See the
[LICENSE][license-link] file for more details.

Projects in the `vendor` folder may have their own LICENSE file.

[golang]: https://golang.org/
[fizzle]: https://github.com/tbogdala/fizzle
[glfw-go]: https://github.com/go-gl/glfw
[mgl]: https://github.com/go-gl/mathgl
[go-gl]: https://github.com/go-gl/glow
[license-link]: https://raw.githubusercontent.com/tbogdala/openvr-go/master/LICENSE
[openvr-git]: https://github.com/ValveSoftware/openvr
