#!/usr/bin/env go-scad

end_cap_sides(2);

left(90);
forward(.5);
right(90);

pendown();
left(45);
// This is a hack but turning arbitrary values into coordinates is the easiest
// way to test them right now!
setpos(heading(), .5);
penup();
