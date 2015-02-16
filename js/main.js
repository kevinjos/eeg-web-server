var canvas = document.getElementById('canvas');
var gl = canvas.getContext("webgl") || canvas.getContext("experimental-webgl");
gl = WebGLDebugUtils.makeDebugContext(gl);
var vertArray;

var transpose = function(a) {
	var t = [];
	for (var i = 0; i < 16; ++i) {
		t[i] = a[(4*i + Math.floor(i/4)) % 16];
	}
	return t;
};

var toFloat32Array = function(a) {
	var f32a = new Float32Array(a.length);
	a.forEach(function(f, i) {
		f32a[i] = f;
	});
	return f32a;
}

var scope = {
	getContext: function()  {
		return gl; 
	},
	
	init: function(data) {
		scope.data = data;
		var rows = data.length;
		var cols = data[0].length;
		var numVerts = scope.getNumVerts(data);
		vertArray = new Float32Array(numVerts * 3);
		var vbo = gl.createBuffer();
		gl.bindBuffer(gl.ARRAY_BUFFER, vbo);
		gl.bufferData(gl.ARRAY_BUFFER, numVerts * vertArray.BYTES_PER_ELEMENT * 3, gl.DYNAMIC_DRAW);
		
		var vertSrc = document.getElementById('vertShader').text;
		var fragSrc = document.getElementById('fragShader').text;
		
		var program = scope.buildShaderProgram(vertSrc, fragSrc);
		var posIndex = gl.getAttribLocation(program, "pos");
		gl.vertexAttribPointer(posIndex, 3, gl.FLOAT, gl.FALSE, vertArray.BYTES_PER_ELEMENT * 3, 0);
		
		gl.enableVertexAttribArray(posIndex);
		
		var modelViewMatLoc = gl.getUniformLocation(program, 'modelViewMat');
		var projMatLoc = gl.getUniformLocation(program, 'projMat');
		gl.useProgram(program);
		var idenMat = new Float32Array(16);
		idenMat[0] = idenMat[5] = idenMat[10] = idenMat[15] = 1;
		var n = 0.01, f = 1000;
		var b = -1, t = 1;
		var l = -1, r = 1;
		var projMat = toFloat32Array(transpose([
			2*n/(r-l), 0,         (r+l)/(r-l),  0,
			0,         2*n/(t-b), (t+b)/(t-b),  0,
			0,         0,         (-n-f)/(f-n), -2*f*n/(f-n),
			0,         0,         -1,           0,
		]));
		
		
		gl.uniformMatrix4fv(modelViewMatLoc, gl.FALSE, idenMat);
		gl.uniformMatrix4fv(projMatLoc, gl.FALSE, projMat);
		
		gl.useProgram(program);
		
		
		scope.buildGraphMesh(data, vbo);
		
		gl.clearColor(0, 0, 1, 1);
		
		gl.clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT);
		gl.drawArrays(gl.TRIANGLES, 0, numVerts);
		
	},
	
	buildShaderProgram: function(vertSrc, fragSrc) {
		var vertShader = gl.createShader(gl.VERTEX_SHADER);
		var fragShader = gl.createShader(gl.FRAGMENT_SHADER);
		gl.shaderSource(vertShader, vertSrc);
		gl.shaderSource(fragShader, fragSrc);
		gl.compileShader(vertShader);
		console.log("vertex shader compilation log:\n",
		gl.getShaderInfoLog(vertShader));
		gl.compileShader(fragShader);
		console.log("fragment shader compilation log:\n",
		gl.getShaderInfoLog(fragShader));
		
		var program = gl.createProgram();
		gl.attachShader(program, vertShader);
		gl.attachShader(program, fragShader);
		
		gl.linkProgram(program);
		
		console.log("linker log:\n",
		gl.getProgramInfoLog(program));
		
		return program;
	},
	
	
	getNumVerts: function(data) {
		var rows = data.length;
		var cols = data[0].length;
		return 12 * (rows-1) * (cols - 1);
	},
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
	buildGraphMesh: function(data, vbo) {
		gl.bindBuffer(gl.ARRAY_BUFFER, vbo);
		var vertArray = new Float32Array(3 * scope.getNumVerts(data));
		var rows = data.length;
		var cols = data[0].length;
		var r = rows - 1;
		var c = cols - 1;
		for (var i = 0; i < r; ++i) {
			for (var j = 0; j < c; ++j) {
				var off = 36*(r*i+j);
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
		gl.bufferData(gl.ARRAY_BUFFER, vertArray.buffer, gl.DYNAMIC_DRAW);
	},
	
};
window.bciVis = scope;
