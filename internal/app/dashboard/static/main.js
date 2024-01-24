import * as d3 from "https://cdn.jsdelivr.net/npm/d3@7/+esm";
import * as topojson from "https://cdn.skypack.dev/topojson-client";

class Map {
    constructor() {
        this.config = {
            map: {
                url: "https://d3js.org/world-50m.v1.json",
                projection: d3.geoNaturalEarth1(),
            },
            aim: [114.176, 22.2782],
            points: {}
        };

        this.width = window.innerWidth;
        this.height = window.innerHeight;
        this.config.map.projection.scale(this.width / 2 / Math.PI).translate([this.width / 2, this.height / 2]);
        this.pathGenerator = d3.geoPath().projection(this.config.map.projection);
    }

    start() {
        this.svg = d3.select("#container")
            .append("svg")
            .attr("width", this.width)
            .attr("height", this.height)
            .style("position", "absolute")
            .style("top", 0)
            .style("left", 0);

        d3.json(this.config.map.url)
            .then(world => {
                this.svg.append("path")
                    .datum(topojson.feature(world, world.objects.land))
                    .attr("fill", "#999")
                    .attr("d", this.pathGenerator);

                this.svg.append("path")
                    .datum(topojson.mesh(world, world.objects.countries, (a, b) => a !== b))
                    .attr("fill", "none")
                    .attr("stroke", "black")
                    .attr("stroke-linejoin", "round")
                    .attr("d", this.pathGenerator);

                this.svg.append("circle")
                    .attr("cx", this.config.map.projection(this.config.aim)[0])
                    .attr("cy", this.config.map.projection(this.config.aim)[1])
                    .attr("r", 3)
                    .attr("fill", "blue");
            })
            .catch(error => console.error(error));
    }

    addPoint(name, location, total) {
        let point = this.config.points[name];
        if (point) {
            point.setTotal(total);
            return;
        }
        point = new Point(name, location, total, this.config.aim, this.svg, this.config.map.projection);
        point.start();
        this.config.points[name] = point;
    }
}

class Point {
    constructor(name, location, count, aim, svg, projection) {
        this.name = name;
        this.location = location
        this.count = count-10;
        this.total = count;
        this.aim = aim;
        this.svg = svg;
        this.projection = projection;
        this.checking = false;
    }

    start() {
        this.svg.append("circle")
            .attr("cx", this.projection(this.location)[0])
            .attr("cy", this.projection(this.location)[1])
            .attr("r", 3)
            .attr("fill", "red");

        const line = d3.line().curve(d3.curveBasis);
        const intermediatePoint = [
            (this.location[0] + this.aim[0]) / 2,
            (this.location[1] + this.aim[1]) / 2 + 10
        ];

        const linePath = this.svg.append("path")
            .datum([
                this.projection(this.location),
                this.projection(intermediatePoint),
                this.projection(this.aim)
            ])
            .attr("fill", "none")
            .attr("stroke", "green")
            .attr("stroke-width", 2)
            .attr("d", line);


        const totalLength = linePath.node().getTotalLength();

        linePath
            .attr("stroke-dasharray", totalLength + " " + totalLength)
            .attr("stroke-dashoffset", totalLength)

        this.line = linePath;

        this.check();
    }

    setTotal(total) {
        this.total = total;
        if (!this.checking) {
            this.check();
        }
    }

    check() {
        this.checking = true;

        if (this.count === this.total) {
            return
        }
        this.count++;

        const totalLength = this.line.node().getTotalLength();

        this.line
            .attr("stroke-dasharray", totalLength + " " + totalLength)
            .attr("stroke-dashoffset", totalLength)
            .transition()
            .duration(1000)
            .ease(d3.easeCubicInOut)
            .attr("stroke-dashoffset", 0)
            .transition()
            .duration(1000)
            .ease(d3.easeCubicInOut)
            .attr("stroke-dashoffset", -totalLength);

        if (this.count === this.total) {
            this.checking = false
            return
        }

        setTimeout(() => {
            this.check()
        }, 3000)
    }
}

const map = new Map();
window.map = map;
map.start();

setInterval(async () => {
    let points = null;
    try {
        const response = await fetch('/api/v1/points');
        if (!response.ok) {
            return;
        }
        const data = await response.json();
        points = data.points;
    } catch (error) {
        console.error(`Fetch error: ${error}`);
    }
    for (const point of points) {
        map.addPoint(point.ip, [point.longitude, point.latitude], point.count);
    }
}, 5000)