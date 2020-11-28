echo('cube([1,2,3]);');
echo('union() {\n\tcube([4,5,6]);\n}');

wrap('union()', function() {
	echo('union() {\n\tcube([4,5,6]);\n}');
});
