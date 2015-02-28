var canvas = document.getElementById('canvas');
var gl = canvas.getContext("webgl") || canvas.getContext("experimental-webgl");
gl = WebGLDebugUtils.makeDebugContext(gl);
var vertArray;

var oldMousePos, mouseDelta = vec2.create();

var cameraPhi = Math.PI/4;
var cameraTheta = 0;

var viewMatLoc;
var geometry;

var vbo;

var scope = {
	getContext: function()  {
		return gl; 
	},
	
	init: function(data) {
		glMatrix.setMatrixArrayType(Float32Array);
		
		canvas.onmousemove = scope.onMouseMove;
		
		gl.enable(gl.DEPTH_TEST);
		
		scope.data = data;
		var rows = data.length;
		var cols = data[0].length;
		
		var vertSrc = document.getElementById('vertShader').text;
		var fragSrc = document.getElementById('fragShader').text;
		
		var program = scope.buildShaderProgram(vertSrc, fragSrc);
		var posIndex = gl.getAttribLocation(program, "pos");
		
		var modelMatLoc = gl.getUniformLocation(program, 'modelMat');
		viewMatLoc = gl.getUniformLocation(program, 'viewMat');
		var projMatLoc = gl.getUniformLocation(program, 'projMat');
		
		gl.useProgram(program);
		var modelMat = mat4.create();
		mat4.translate(modelMat, modelMat, vec3.fromValues(-0.5, 0, -0.5));
		mat4.scale(modelMat, modelMat, vec3.fromValues(1, 1/4, 1));
		var viewMat = mat4.create()
		mat4.translate(viewMat, viewMat, vec3.fromValues(0, 0, -3));
		mat4.rotateX(viewMat, viewMat, Math.PI/2);
		var n = 0.01, f = 1000;
		var b = -0.01, t = 0.01;
		var l = -0.01, r = 0.01;
		var projMat = mat4.frustum(mat4.create(), l, r, b, t, n, f);
		
		
		gl.uniformMatrix4fv(modelMatLoc, false, modelMat);
		gl.uniformMatrix4fv(viewMatLoc, false, viewMat);
		gl.uniformMatrix4fv(projMatLoc, false, projMat);
		
		gl.useProgram(program);
		
		geometry = makeGraph(data);
		vbo = makeVbo(geometry, gl.DYNAMIC_DRAW);
		gl.bindBuffer(gl.ARRAY_BUFFER, vbo);
		
		gl.vertexAttribPointer(posIndex, 3, gl.FLOAT, false,  4 * 3, 0);
		
		gl.enableVertexAttribArray(posIndex);
		
		scope.onStep();
		//var vbo = scope.makeGraph(data);
		//gl.bindBuffer(gl.ARRAY_BUFFER, vbo);
		
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
	
	onMouseMove: function(event) {
		if (!oldMousePos) {
			oldMousePos = vec2.fromValues(event.x, event.y);
			mouseDelta = vec2.create();
			return;
		}
		vec2.set(mouseDelta, event.x - oldMousePos[0], event.y - oldMousePos[1]);
		vec2.set(oldMousePos, event.x, event.y);
	},
	
	updateData: function(newDatum) {
		scope.data.splice(0, 0, newDatum);
		scope.data.splice(scope.data.length - 1, 1);
		geometry = makeGraph(scope.data);
		makeVbo(geometry, gl.DYNAMIC_DRAW, vbo);
		scope.render();
	},
	
	onStep: function() {
		if (mouseDelta[0] === 0 && mouseDelta[1] === 0) {
			requestAnimationFrame(scope.onStep);
			return;
		}
		cameraTheta += mouseDelta[0] / 30;
		cameraPhi += mouseDelta[1] / 30;
		if (cameraPhi > Math.PI/2) cameraPhi = Math.PI/2;
		if (cameraPhi < -Math.PI/2) cameraPhi = -Math.PI/2;
		
		var viewMat = mat4.create();
		
		mat4.translate(viewMat, viewMat, vec3.fromValues(0, 0, -2));
		
		mat4.rotateX(viewMat, viewMat, cameraPhi);
		mat4.rotateY(viewMat, viewMat, cameraTheta);
		
		
		
		
		gl.uniformMatrix4fv(viewMatLoc, false, viewMat);
		
		scope.render()
		
		vec2.set(mouseDelta, 0, 0);
		requestAnimationFrame(scope.onStep);
	},
	
	render: function() {
		gl.clearColor(0, 0, 0, 1);
		gl.clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT);
		gl.drawArrays(gl.TRIANGLES, 0, geometry.length/3);
	},
	
};
window.bciVis = scope;

var zeroData = [];
for (var i = 0; i < 20; ++i) {
	zeroData.push([]);
	for (var j = 0; j < 25; ++j) {
		zeroData[i].push(0);
	}
}

scope.init(zeroData);
