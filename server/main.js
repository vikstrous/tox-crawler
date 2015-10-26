var redraw;
(function() {
    d3.select(window).on("resize", throttle);

    var zoom = d3.behavior.zoom()
        .scaleExtent([1, 9])
        .on("zoom", move);

    var width = document.getElementById('container').offsetWidth;
    var height = width / 2;

    var topo, projection, path, svg2, g;

    var tooltip = d3.select("#container").append("div").attr("class", "tooltip hidden");

    setup(width, height);

    function setup(width, height) {
        projection = d3.geo.mercator()
            .translate([(width / 2), (height / 2)])
            .scale(width / 2 / Math.PI);

        path = d3.geo.path().projection(projection);

        svg2 = d3.select("#container").append("svg")
            .attr("width", width)
            .attr("height", height)
            .call(zoom)
            .append("g");

        g = svg2.append("g");
    }

    d3.json("data/world-topo-min.json", function(error, world) {
        var countries = topojson.feature(world, world.objects.countries).features;
        topo = countries;
        draw(topo);
    });

    function draw(topo) {
        var country = g.selectAll(".country").data(topo);

        country.enter().insert("path")
            .attr("class", "country")
            .attr("d", path)
            .attr("id", function(d, i) {
                return d.id;
            })
            .style("fill", function(d, i) {
                return country_color(d.properties.name);
            });

        //offsets for tooltips
        var offsetL = document.getElementById('container').offsetLeft + 20;
        var offsetT = document.getElementById('container').offsetTop + 10;

        //tooltips
        country
            .on("mousemove", function(d, i) {
                var mouse = d3.mouse(svg2.node()).map(function(d) {
                    return parseInt(d);
                });

                tooltip.classed("hidden", false)
                    .attr("style", "left:" + (mouse[0] + offsetL) + "px;top:" + (mouse[1] + offsetT) + "px")
                    .html(
                        name_correction(d.properties.name) + ": " + get_count(d.properties.name)
                    );

            })
            .on("mouseout", function(d, i) {
                tooltip.classed("hidden", true);
            });
    }

    redraw = function() {
        width = document.getElementById('container').offsetWidth;
        height = width / 2;
        d3.select('#container svg').remove();
        setup(width, height);
        draw(topo);
    }

    function move() {
        var t = d3.event.translate;
        var s = d3.event.scale;
        zscale = s;
        var h = height / 4;


        t[0] = Math.min(
            (width / height) * (s - 1),
            Math.max(width * (1 - s), t[0])
        );

        t[1] = Math.min(
            h * (s - 1) + h * s,
            Math.max(height * (1 - s) - h * s, t[1])
        );

        zoom.translate(t);
        g.attr("transform", "translate(" + t + ")scale(" + s + ")");

        //adjust the country hover stroke width based on zoom level
        d3.selectAll(".country").style("stroke-width", 1.5 / s);

    }

})();

var throttleTimer;

function throttle() {
    window.clearTimeout(throttleTimer);
    throttleTimer = window.setTimeout(function() {
        redraw();
        render_graph();
    }, 200);
}

// handles values up to 255
function to_hex(value){
    value = Math.floor(value);
    var upper = Math.floor(value / 16);
    var lower = value % 16;
    return hex[upper] + hex[lower];
}

function country_color(name) {
    var count = get_count(name);
    if (country_counts_max) {
    var percent = count / country_counts_max;
    // give the space a bit of an exponential curve
    percent = Math.log(Math.floor(percent * 1023) + 1) / Math.log(2) / 10
    var blue = Math.floor(percent * 255);
    var green = Math.floor(blue)/2;
    return '#00' + to_hex(green) + to_hex(blue);
    } else {
        return '#000000';
    }
}

function get_count(name) {
    name = name_correction(name);
    if (country_counts && country_counts[name]) {
        return country_counts[name];
    }
    return 0;
}
function name_correction(name) {
    name = name.split(',')[0];
    if (name_corrections[name] !== undefined) {
        return name_corrections[name]
    }
    return name
}
var name_corrections = {
    'Russian Federation': 'Russia',
    'Korea': 'Republic of Korea',
    'Lithuania': 'Republic of Lithuania',
    'Moldova': 'Republic of Moldova',
    'Slovakia': 'Slovak Republic'
}



var hex = '0123456789abcdef';
var country_counts;
var country_counts_max;
var svg1;
var data;
var render_graph;
$(function() {
    // build the path to the websocket
    var loc = window.location,
        new_uri;
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
        var parsed = JSON.parse(e.data);
        console.log(parsed);
        country_counts = parsed.country_counts;
        var max = 0;
        for (var k in country_counts) {
            var v = country_counts[k];
            if (v > max) {
                max = v;
            }
        }
        country_counts_max = max;
        data = parsed.numbers_list;
        render_graph();
        redraw();
    };

    var margin = {
            top: 20,
            right: 20,
            bottom: 30,
            left: 50
        };


    render_graph = function() {
        var width = document.getElementById('container').offsetWidth - margin.left - margin.right;
        var height = width / 2 - margin.top - margin.bottom;

        var parseDate = d3.time.format.iso.parse;

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
            .x(function(d) {
                return x(d.date);
            })
            .y(function(d) {
                return y(d.close);
            });

        if (svg1) {
            $(svg1[0]).parent().remove()
        }
        svg1 = d3.select("#graph").append("svg")
            .attr("width", width + margin.left + margin.right)
            .attr("height", height + margin.top + margin.bottom)
            .append("g")
            .attr("transform", "translate(" + margin.left + "," + margin.top + ")");

        data.forEach(function(d) {
            d.date = parseDate(d.time);
            d.close = +d.number;
        });

        x.domain(d3.extent(data, function(d) {
            return d.date;
        }));
        y.domain(d3.extent(data, function(d) {
            return d.close;
        }));

        svg1.append("g")
            .attr("class", "x axis")
            .attr("transform", "translate(0," + height + ")")
            .call(xAxis);

        svg1.append("g")
            .attr("class", "y axis")
            .call(yAxis)
            .append("text")
            .attr("transform", "rotate(-90)")
            .attr("y", 10)
            .attr("dy", ".71em")
            .style("text-anchor", "end")
            .text("Nodes");

        svg1.append("path")
            .datum(data)
            .attr("class", "line")
            .attr("d", line);
    }
});
