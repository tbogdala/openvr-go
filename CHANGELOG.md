Version v0.4.2
==============

* MISC: Changed the `vendor` directory to `vendored` to support including this library with Go's
  `dep` tool, which currently will drop that vendor flag.

Version v0.4.1
==============

* BUG: Build fixes for Linux systems.

Version v0.4.0
==============

* NEW: ICompositor support for GetFrameTimeRemaining() and GetFrameTiming().

Version v0.3.0
==============

* APIBREAK: Changes were made to support OpenVR 1.0.5 upstream. Updated binaries.
  Removed linux32 from lib & bin. Reviewed enumerations and brought some sets into
  conformity of the naming convention.

* MISC: Switched to using github.com/tbogdala/fizzle's built in shaders for samples.

* MISC: Switched to Mathgl for vectors instead of github.com/tbogdala/glider's.

* MISC: Switched to using fizzle's Material object in examples.

Version v0.2.0
==============

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
