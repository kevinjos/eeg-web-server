var makeCube = function(size) {
	var l = size / 2;
	/*
	    y
	    |
	    e------f
	   /|     /|
	  h------g |
	  | a----|-b---x
	  |/     |/
	  d------c
	 /
	z

	*/
	
	var a = [-l, -l, -l];
	var b = [+l, -l, -l];
	var c = [+l, -l, +l];
	var d = [-l, -l, +l];
	var e = [-l, +l, -l];
	var f = [+l, +l, -l];
	var g = [+l, +l, +l];
	var h = [-l, +l, +l];
	
	var quadverts = function(a, b, c, d) {
		return [].concat(a, b, d, b, c, d);
	}
	
	var verts = [].concat(
		quadverts(a, b, c, d),
		quadverts(e, h, g, f),
		quadverts(d, c, g, h),
		quadverts(b, a, e, f),
		quadverts(c, b, f, g),
		quadverts(e, a, d, h)
	);
	return new Float32Array(verts);
}

var makeGraph = function(data) {
				
/*
	.-----X--->
	|
	Z  a---b---+-
	|  |\4/|\ /|\
	|  |1e3| + |
	V  |/2\|/ \|/ ...
	   c---d---+-
	   |\ /|\ /|\
	      .
	   	  .
	      .
		
*/
	var rows = data.length;
	var cols = data[0].length;
	var r = rows - 1;
	var c = cols - 1;
	var numVerts = r*c*12;
	var vertArray = new Float32Array(3 * numVerts);
	for (var i = 0; i < r; ++i) {
		for (var j = 0; j < c; ++j) {
			var off = 36*(c*i+j);
			var a_x = j/c;
			var b_x = (j+1)/c;
			var c_x = a_x;
			var d_x = b_x;
			var a_z = i/r;
			var c_z = (i+1)/r;
			var b_z = a_z;
			var d_z = c_z;
			var e_x = (a_x+b_x)/2
			var e_z = (a_z+c_z)/2;
			var a_y = data[i][j];
			var b_y = data[i][j+1];
			var c_y = data[i+1][j];
			var d_y = data[i+1][j+1];
			var e_y = (a_y+b_y+c_y+d_y)/4;
			
			vertArray[off + 0] = a_x;
			vertArray[off + 1] = a_y;
			vertArray[off + 2] = a_z;
			vertArray[off + 3] = c_x;
			vertArray[off + 4] = c_y;
			vertArray[off + 5] = c_z;
			vertArray[off + 6] = e_x;
			vertArray[off + 7] = e_y;
			vertArray[off + 8] = e_z;
			
			vertArray[off + 9] = c_x;
			vertArray[off +10] = c_y;
			vertArray[off +11] = c_z;
			vertArray[off +12] = d_x;
			vertArray[off +13] = d_y;
			vertArray[off +14] = d_z;
			vertArray[off +15] = e_x;
			vertArray[off +16] = e_y;
			vertArray[off +17] = e_z;
			
			vertArray[off +18] = d_x;
			vertArray[off +19] = d_y;
			vertArray[off +20] = d_z;
			vertArray[off +21] = b_x;
			vertArray[off +22] = b_y;
			vertArray[off +23] = b_z;
			vertArray[off +24] = e_x;
			vertArray[off +25] = e_y;
			vertArray[off +26] = e_z;
			
			vertArray[off +27] = b_x;
			vertArray[off +28] = b_y;
			vertArray[off +29] = b_z;
			vertArray[off +30] = a_x;
			vertArray[off +31] = a_y;
			vertArray[off +32] = a_z;
			vertArray[off +33] = e_x;
			vertArray[off +34] = e_y;
			vertArray[off +35] = e_z;
			
		}
	}
	return vertArray;
};



var makeVbo = function(geometry, mode, vbo){
	mode = mode || gl.STATIC_DRAW;
	if (!vbo) {
		vbo = gl.createBuffer();
	}
	gl.bindBuffer(gl.ARRAY_BUFFER, vbo);
	gl.bufferData(gl.ARRAY_BUFFER, geometry, mode);
	return vbo;
}
