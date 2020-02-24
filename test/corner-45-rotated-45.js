#!/usr/bin/env go-scad

end_cap_sides(6);

left(90);
forward(.5);
right(45);

pendown();
forward(5);
left(45);
forward(5);
penup();
