#!/usr/bin/env go-scad

end_cap_sides(6);

left(45);

pendown();
forward(3);
setpos(3, 0);
forward(3);
setpos(6, 0);
setpos(9, 0);
penup();
