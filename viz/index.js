import * as d3 from "https://cdn.jsdelivr.net/npm/d3@7/+esm";

const data = await (await fetch("/data.json")).json()

graph(data)

function graph(data) {
	// Specify the chart’s dimensions.
	const width = window.innerWidth - 5;
	const height = window.innerHeight - 5;
	let cx = width * 0.53; // adjust as needed to fit
	let cy = height * 0.5; // adjust as needed to fit
	const radius = Math.min(width, height) / 2 - 30;

	// Create a radial tree layout. The layout’s first dimension (x)
	// is the angle, while the second (y) is the radius.
	const tree = d3.tree()
		.size([2 * Math.PI, radius])
		.separation((a, b) => (a.parent == b.parent ? 1 : 2) / a.depth);

	// Sort the tree and apply the layout.
	const root = tree(d3.hierarchy(data)
		.sort((a, b) => d3.ascending(a.data.name, b.data.name)));

	// Creates the SVG container.
	const svg = d3.create("svg")
		.attr("width", width)
		.attr("height", height)
		.attr("viewBox", [-cx, -cy, width, height])

	const g = svg.append("g")

	const zoom = d3.zoom()
		.scaleExtent([0.5, 5])
		.on("zoom", (e) => {
			g.attr('transform', e.transform)
		})

	d3.select(window).on("keydown", (e) => {
		const step = 0.2
		let transform = d3.zoomTransform(svg.node())
		let newScale = transform.k

		switch (e.key) {
			case "+":
			case "=":
				newScale = transform.k * (1 + step);
				break
			case "-":
			case "_":
				newScale = transform.k * (1 - step);
				break
			default:
				return
		}

		svg.transition()
			.duration(0)
			.call(zoom.scaleTo, newScale);
	})

	svg.call(zoom)

	// Append links.
	g.append("g")
		.attr("fill", "none")
		.attr("stroke", "#555")
		.attr("stroke-opacity", 0.4)
		.attr("stroke-width", 1.5)
		.selectAll()
		.data(root.links())
		.join("path")
		.attr("d", d3.linkRadial()
			.angle(d => d.x)
			.radius(d => d.y));

	// Append nodes.
	g.append("g")
		.selectAll()
		.data(root.descendants())
		.join("circle")
		.attr("transform", d => `rotate(${d.x * 180 / Math.PI - 90}) translate(${d.y},0)`)
		.attr("fill", d => d.children ? "#555" : "#999")
		.attr("r", 2.5);

	// Append labels.
	g.append("g")
		.attr("stroke-linejoin", "round")
		.attr("stroke-width", 3)
		.selectAll()
		.data(root.descendants())
		.join("text")
		.attr("transform", d => `rotate(${d.x * 180 / Math.PI - 90}) translate(${d.y},0) rotate(${d.x >= Math.PI ? 180 : 0})`)
		.attr("dy", "0.31em")
		.attr("x", d => d.x < Math.PI === !d.children ? 6 : -6)
		.attr("text-anchor", d => d.x < Math.PI === !d.children ? "start" : "end")
		.attr("paint-order", "stroke")
		.attr("stroke", "white")
		.attr("fill", "currentColor")
		.text(d => d.data.name);

	document.body.append(svg.node())
}
