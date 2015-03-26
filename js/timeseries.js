var rawScope = {
  makechart : function(chartid) {
    var n = 250;
    var data = d3.range(n).map(function() {return 0});

    var margin = {top: 20, right: 30, bottom: 20, left: 40},
        width = 380 - margin.left - margin.right,
        height = 190 - margin.top - margin.bottom;

    var x = d3.scale.linear()
        .domain([0, n - 1])
        .range([0, width]);

    var xAxis = d3.svg.axis()
        .scale(x)
        .ticks(0)
        .orient("bottom");

    var y = d3.scale.linear()
        .domain([-10.0, 10.0])
        .range([height, 0])
        .clamp(true);

    var yAxis = d3.svg.axis()
        .scale(y)
        .orient("left");

    var line = d3.svg.line()
      .x(function(d, i) { return x(i); })
      .y(function(d, i) { return y(d); }); 

    var chart = d3.select("#" + chartid)
        .attr("width", width + margin.left + margin.right)
        .attr("height", height + margin.top + margin.bottom)
        .style("color", "black")
        .style("background-color", "white")
      .append("g")
        .attr("transform", "translate(" + margin.left + "," + margin.top + ")");

    chart.append("g")
        .attr("class", "x axis")
        .attr("transform", "translate(0," + height/2 + ")")
        .call(xAxis);

    chart.append("g")
        .attr("class", "y axis")
        .call(yAxis);

    chart.append("defs").append("clipPath")
      .attr("id", "clip")
      .append("rect")
      .attr("width", width)
      .attr("height", height); 

    var path = chart.append("g")
      .attr("clip-path", "url(#clip)")
      .append("path")
      .datum(data)
      .attr("class", "line")
      .attr("d", line); 

    return function(d) {
      for (i = 0; i < d.length; i++) {
          data.push(d[i]);
          data.shift();
      }
      // redraw the line, and slide it to the left
      path
      .attr("d", line)
      .attr("transform", null)
      .transition()
        //.duration(20)
        //.ease("linear")
        //.attr("transform", "translate(" + x(-1) + ",0)")
    }
  },


	tick : function() {
		// push a new data point onto the back
		data.push(random());
		// redraw the line, and slide it to the left
		path
		.attr("d", line)
		.attr("transform", null)
		.transition()
		//.duration(500)
		//.ease("linear")
		//.attr("transform", "translate(" + x(-1) + ",0)")
		//.each("end", scope.tick);
		// pop the old data point off the front
		data.shift();
	}, 
};

window.updaterawchan1 = rawScope.makechart("timeserieschan1")
window.updaterawchan2 = rawScope.makechart("timeserieschan2")
window.updaterawchan3 = rawScope.makechart("timeserieschan3")
window.updaterawchan4 = rawScope.makechart("timeserieschan4")
window.updaterawchan5 = rawScope.makechart("timeserieschan5")
window.updaterawchan6 = rawScope.makechart("timeserieschan6")
window.updaterawchan7 = rawScope.makechart("timeserieschan7")
window.updaterawchan8 = rawScope.makechart("timeserieschan8")
window.timeseries = rawScope;

