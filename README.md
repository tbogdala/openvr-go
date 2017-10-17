Openvr-go v0.4.1
================

Openvr-go is an [Go][golang] programming language wrapper for the [OpenVR SDK][openvr-git]
published by Valve for VR hardware.

This package is currently synced up to v1.0.10 of OpenVR.

![voxels_ss][voxels_ss]

UNDER CONSTRUCTION
==================

At present, it is very much in an alpha stage with new development happening to
complete the API exposed by the OpenVR SDK.

Requirements
------------

* [Mathgl][mgl] - for 3d math

The wrapper library itself doesn't have any dependencies besides [Mathgl][mgl].
The `connectiontest` sample in the `examples` folder also doesn't have any
additional dependencies.

The other samples are graphical and use the following libraries, though they are
not imported by the core openvr-go module itself:

* [GLFW][glfw-go] (v3.1) - creating windows and providing the OpenGL context
* [Fizzle][fizzle] (v0.2.0) - provides the graphics engine
* [Go GL][go-gl] - provides the backend implementation of OpenGL for [Fizzle][fizzle].

Note: At present, some examples might required the development branch of [Fizzle][fizzle].
You'll have to manually git checkout the `development` branch to compile these.

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
`vendored/openvr/bin/<platform>` will either need to be copied into each example directory
being built or it will need to be accessible system wide.

Each sample can be built by going to that directory in a shell and executing
a `go build` command. For example:

```bash
cd $GOPATH/src/github.com/tbogdala/openvr-go/examples/basiccube
go build
cp ../../vendored/openvr/bin/win64/openvr_api.dll .
./basiccube.exe
```

Current Features
----------------

Partial implementation of the following interfaces:

* IVRSystem
* IVRCompositor
* IVRRenderModels


Implementation Notes
--------------------

Some minor patches have been applied to the vendored openvr library version to
better support linux.


LICENSE
=======

Original source code in openvr-go is released under the BSD license. See the
[LICENSE][license-link] file for more details.

Projects in the `vendor` folder may have their own LICENSE file.

The MTCORE32px texture pack files in `examples/voxels/assets/textures` are licensed
CC BY-SA 3.0 by celeron55, Perttu Ahola.
https://github.com/Napiophelios/MTCORE32px

[golang]: https://golang.org/
[fizzle]: https://github.com/tbogdala/fizzle
[glfw-go]: https://github.com/go-gl/glfw
[mgl]: https://github.com/go-gl/mathgl
[go-gl]: https://github.com/go-gl/glow
[license-link]: https://raw.githubusercontent.com/tbogdala/openvr-go/master/LICENSE
[openvr-git]: https://github.com/ValveSoftware/openvr
[basiccube_ss]: https://raw.githubusercontent.com/tbogdala/openvr-go/master/examples/screenshots/example-basiccube.jpg
[voxels_ss]: https://github.com/tbogdala/openvr-go/blob/development/examples/screenshots/example-voxels.jpg?raw=true
