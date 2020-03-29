#!/usr/bin/env go-scad

end_cap_sides(6);

left(90);
forward(.5);
right(90);

pendown();
setpos(5, .5);
setpos(5 + 5*Math.sqrt(2)/2, .5 + 5*Math.sqrt(2)/2);
penup();
