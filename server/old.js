var redraw;
(function(){
d3.select(window).on("resize", throttle);

var zoom = d3.behavior.zoom()
    .scaleExtent([1, 9])
    .on("zoom", move);


var width = document.getElementById('container').offsetWidth;
var height = width / 2;

var topo,projection,path,svg2,g;

var graticule = d3.geo.graticule();

var tooltip = d3.select("#container").append("div").attr("class", "tooltip hidden");

setup(width,height);

function setup(width,height){
  projection = d3.geo.mercator()
    .translate([(width/2), (height/2)])
    .scale( width / 2 / Math.PI);

  path = d3.geo.path().projection(projection);

  svg2 = d3.select("#container").append("svg")
      .attr("width", width)
      .attr("height", height)
      .call(zoom)
      //.on("click", click)
      .append("g");

  g = svg2.append("g")
         .on("click", click);

}

d3.json("data/world-topo-min.json", function(error, world) {

  var countries = topojson.feature(world, world.objects.countries).features;

  topo = countries;
  draw(topo);

});

function draw(topo) {

  svg2.append("path")
     .datum(graticule)
     .attr("class", "graticule")
     .attr("d", path);


  g.append("path")
   .datum({type: "LineString", coordinates: [[-180, 0], [-90, 0], [0, 0], [90, 0], [180, 0]]})
   .attr("class", "equator")
   .attr("d", path);


  var country = g.selectAll(".country").data(topo);

  country.enter().insert("path")
      .attr("class", "country")
      .attr("d", path)
      .attr("id", function(d,i) { return d.id; })
      //.attr("title", function(d,i) { return name_correction(d.properties.name); })
      .style("fill", function(d, i) {
          if (country_counts) {
            var count = country_counts[name_correction(d.properties.name)];
                if (count !== undefined) {
                    var percent = count / country_counts_max;
                    // give the space a bit of an exponential curve
                    percent = Math.log(Math.floor(percent*1023)+1)/Math.log(2)/10
                    var value = Math.floor(percent * 255);
                    var upper = Math.floor(value / 16);
                    var lower = value % 16;
                    if (d.properties.name == 'United States') {
                        console.log('#0000' + hex[upper] + hex[lower])
                    }
              return '#0000' + hex[upper] + hex[lower];
                    } else {
              return '#000000';
                    }
          } else {
              return d.properties.color;
          }
      });

  //offsets for tooltips
  var offsetL = document.getElementById('container').offsetLeft+20;
  var offsetT = document.getElementById('container').offsetTop+10;

  //tooltips
  country
    .on("mousemove", function(d,i) {

      var mouse = d3.mouse(svg2.node()).map( function(d) { return parseInt(d); } );

      tooltip.classed("hidden", false)
             .attr("style", "left:"+(mouse[0]+offsetL)+"px;top:"+(mouse[1]+offsetT)+"px")
             .html(
                     d.properties.name + ": " + country_counts[name_correction(d.properties.name)]
                  );

      })
      .on("mouseout",  function(d,i) {
        tooltip.classed("hidden", true);
      }); 


  //EXAMPLE: adding some capitals from external CSV file
  //d3.csv("data/country-capitals.csv", function(err, capitals) {

  //  capitals.forEach(function(i){
  //    addpoint(i.CapitalLongitude, i.CapitalLatitude, i.CapitalName );
  //  });

  //});

}


redraw = function() {
  width = document.getElementById('container').offsetWidth;
  height = width / 2;
  d3.select('#container svg').remove();
  setup(width,height);
  draw(topo);
}


function move() {

  var t = d3.event.translate;
  var s = d3.event.scale; 
  zscale = s;
  var h = height/4;


  t[0] = Math.min(
    (width/height)  * (s - 1), 
    Math.max( width * (1 - s), t[0] )
  );

  t[1] = Math.min(
    h * (s - 1) + h * s, 
    Math.max(height  * (1 - s) - h * s, t[1])
  );

  zoom.translate(t);
  g.attr("transform", "translate(" + t + ")scale(" + s + ")");

  //adjust the country hover stroke width based on zoom level
  d3.selectAll(".country").style("stroke-width", 1.5 / s);

}



var throttleTimer;
function throttle() {
  window.clearTimeout(throttleTimer);
    throttleTimer = window.setTimeout(function() {
      redraw();
    }, 200);
}


//geo translation on mouse click in map
function click() {
  var latlon = projection.invert(d3.mouse(this));
  console.log(latlon);
}


//function to add points and text to the map (used in plotting capitals)
function addpoint(lat,lon,text) {

  var gpoint = g.append("g").attr("class", "gpoint");
  var x = projection([lat,lon])[0];
  var y = projection([lat,lon])[1];

  gpoint.append("svg:circle")
        .attr("cx", x)
        .attr("cy", y)
        .attr("class","point")
        .attr("r", 1.5);

  //conditional in case a point has no associated text
  if(text.length>0){

    gpoint.append("text")
          .attr("x", x+2)
          .attr("y", y+2)
          .attr("class","text")
          .text(text);
  }

}
})();




function name_correction(name) {
    if (name_corrections[name] !== undefined) {
        return name_corrections[name]
    }
    name = name.split(',')[0];
    return name
}
var name_corrections = {
    'Russian Federation': 'Russia'
}



var hex = '0123456789abcdef';
var country_counts;
var country_counts_max;
var svg1;
    $(function() {
        // build the path to the websocket
        var loc = window.location, new_uri;
        if (loc.protocol === "https:") {
            new_uri = "wss:";
        } else {
            new_uri = "ws:";
        }
        new_uri += "//" + loc.host;
        new_uri += loc.pathname + "stats";

        // connect
        var ws = new WebSocket(new_uri);
        // handle new data
        ws.onmessage = function(e) {
          var data = JSON.parse(e.data);
          console.log(data);
          country_counts = data.country_counts;
          var max = 0;
          for (var k in country_counts) {
              var v = country_counts[k];
              if (v > max) {
                  max = v;
              }
          }
          country_counts_max = max;
          render_graph(data.numbers_list);
            redraw();
        };

        // allow forced refresh
        $('#sendBtn').click(function(){
              ws.send("refresh");
        });

var margin = {top: 20, right: 20, bottom: 30, left: 50},
    width = 960 - margin.left - margin.right,
    height = 500 - margin.top - margin.bottom;

//var parseDate = Date
var parseDate = d3.time.format.iso.parse;
//var parseDate = d3.time.format("%d-%b-%y").parse;

var x = d3.time.scale()
    .range([0, width]);

var y = d3.scale.linear()
    .range([height, 0]);

var xAxis = d3.svg.axis()
    .scale(x)
    .orient("bottom");

var yAxis = d3.svg.axis()
    .scale(y)
    .orient("left");

var line = d3.svg.line()
    .x(function(d) { return x(d.date); })
    .y(function(d) { return y(d.close); });

function render_graph(data) {
if (svg1) {
    $(svg1[0]).parent().remove()
}
svg1 = d3.select("#graph").append("svg")
    .attr("width", width + margin.left + margin.right)
    .attr("height", height + margin.top + margin.bottom)
    .append("g")
    .attr("transform", "translate(" + margin.left + "," + margin.top + ")");
    if (data === null) {
        data = [
       // {number: 2253,
       //     time: "2015-10-23T01:31:45.452030533-07:00"},
       // {number: 2222,
       //     time: "2015-10-23T01:32:30.38162021-07:00"},
       // {number: 2225,
       //     time: "2015-10-23T01:33:20.942007336-07:00"}
        ];
    }
//    data = [
//    { time:"1-May-12", number:	2253},
//    {time:"30-Apr-12", number:	2222},
//    {time:"27-Apr-12", number:	2225}
//];
  data.forEach(function(d) {
    d.date = parseDate(d.time);
    d.close = +d.number;
  });

  x.domain(d3.extent(data, function(d) { return d.date; }));
  y.domain(d3.extent(data, function(d) { return d.close; }));

  svg1.append("g")
      .attr("class", "x axis")
      .attr("transform", "translate(0," + height + ")")
      .call(xAxis);

  svg1.append("g")
      .attr("class", "y axis")
      .call(yAxis)
    .append("text")
      .attr("transform", "rotate(-90)")
      .attr("y", -20)
      .attr("dy", ".71em")
      .style("text-anchor", "end")
      .text("Nodes");

  svg1.append("path")
      .datum(data)
      .attr("class", "line")
      .attr("d", line);
}

    });


