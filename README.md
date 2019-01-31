# go-scad

This is a tool to aid in construction of complex shapes using
[OpenSCAD](https://www.openscad.org/).
I wrote it because it's difficult to make objects with rounded corners and
generally smooth transitions between features using OpenSCAD, but otherwise, it
is the best way I've found to define 3D models using code.

For now, this tool is 2D only, with the intent to expand to 3D later.

**Input: `file.js`**
<br>
This file can use standard JavaScript syntax as well as a library similar to
[turtle graphics for Python](https://docs.python.org/3.3/library/turtle.html)
to draw basic 2-dimensional shapes.

**Output: `file.js.scad`**
<br>
Contains OpenSCAD code with
[`polygon()`](https://en.wikibooks.org/wiki/OpenSCAD_User_Manual/Using_the_2D_Subsystem#polygon)
statements.  You can import this code into other OpenSCAD files and use the
[extrusion functions](https://en.wikibooks.org/wiki/OpenSCAD_User_Manual/2D_to_3D_Extrusion)
to turn them into 3D shapes.
