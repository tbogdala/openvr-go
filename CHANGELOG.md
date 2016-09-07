Changes since v0.2.0
====================

* APIBREAK: Library now uses github.com/go-gl/mathgl/mgl32 for Vector and
  Matrix types where there used to be local definitions.

* NEW: IChaperone support.

* NEW: More IRenderModel functions supported.

* NEW: Voxel engine sample in `examples/voxels`! You start at the edge of a play
  area and can teleport short distances by pulling the trigger on a controller
  and pointing to land.

  This example uses several additional libraries from github.com/tbogdala including
  glider, cubez, and fizzle.

  The shaders used in this sample are based on an older
  ADS-type shader in github.com/tbogdala/fizzle ... and eventually should be
  updated.

* NEW: refactored code from `examples/basiccube` to `openvr-go/util/fizzlevr` which
  makes it easier to start new applications using the github.com/tbogdala/fizzle
  graphics library.

* BUG: added binaries in `vendor/openvr/bin` for win32 and win64 that were missing.

* MISC: Added IChaperone play area size printing to `examples/connectiontest`.

* MISC: Better screenshot.
