#!/usr/bin/env go-scad

end_cap_sides(6);

left(90);
forward(.5);
right(90);

pendown();
forward(5);
left(90);
forward(5);
penup();
