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
		glMatrix.setMatrixArrayType(Float32Array);
		
		scope.data = data;
		var rows = data.length;
		var cols = data[0].length;
		
		var vertSrc = document.getElementById('vertShader').text;
		var fragSrc = document.getElementById('fragShader').text;
		
		var program = scope.buildShaderProgram(vertSrc, fragSrc);
		var posIndex = gl.getAttribLocation(program, "pos");
		
		var modelMatLoc = gl.getUniformLocation(program, 'modelMat');
		var viewMatLoc = gl.getUniformLocation(program, 'viewMat');
		var projMatLoc = gl.getUniformLocation(program, 'projMat');
		
		gl.useProgram(program);
		var modelMat = mat4.create();
		var viewMat = mat4.create()
		mat4.translate(viewMat, viewMat, vec3.fromValues(0, 0, -3));
		mat4.rotateX(viewMat, viewMat, Math.PI/2);
		var n = 0.01, f = 1000;
		var b = -0.001, t = 0.001;
		var l = -0.01, r = 0.01;
		var projMat = mat4.frustum(mat4.create(), l, r, b, t, n, f);
		
		
		gl.uniformMatrix4fv(modelMatLoc, false, modelMat);
		gl.uniformMatrix4fv(viewMatLoc, false, viewMat);
		gl.uniformMatrix4fv(projMatLoc, false, projMat);
		
		gl.useProgram(program);
		
		var geometry = makeGraph(data);
		var vbo = makeVbo(geometry);
		gl.bindBuffer(gl.ARRAY_BUFFER, vbo);
		
		gl.vertexAttribPointer(posIndex, 3, gl.FLOAT, false,  4 * 3, 0);
		
		gl.enableVertexAttribArray(posIndex);
		
		//var vbo = scope.makeGraph(data);
		//gl.bindBuffer(gl.ARRAY_BUFFER, vbo);
		
		gl.clearColor(0, 0, 1, 1);
		
		gl.clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT);
		gl.drawArrays(gl.TRIANGLES, 0, geometry.length/3);
		gl.finish();
		console.log('done');
		
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
	
};
window.bciVis = scope;

var zeroData = [];
for (var i = 0; i < 100; ++i) {
	zeroData.push([]);
	for (var j = 0; j < 8; ++j) {
		zeroData[i].push((i+j) % 2);
	}
}

scope.init(zeroData);
