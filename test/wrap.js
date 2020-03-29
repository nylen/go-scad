#!/usr/bin/env go-scad

wrap('linear_extrude(height = 3)', function() {
	end_cap_sides(6);
	pendown();
	penup();
})
