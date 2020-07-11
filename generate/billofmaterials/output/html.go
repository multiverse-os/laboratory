package output

import (
	"fmt"
	"io"
	"text/template"

	"github.com/Rhymond/go-money"
	"github.com/bom-maker/core"
	"github.com/bom-maker/mouser/model"
)

var (
	_ Writer = (*HTML)(nil)
)

// HTML outputs as HTML format
type HTML struct {
	Parts []core.UberPart
	Title string
}

func (o *HTML) Write(w io.Writer) error {
	// Icons from https://www.flaticon.com/packs/semiconductor-4?word=chip

	funcMap := template.FuncMap{
		"GetPrice": func(part core.UberPart) string {
			pb := part.GetUnitPrice(part.Quantity)
			amount := model.GetAPIFloatFromString(pb.Price)
			//grapheme := money.New(0, pb.Currency).Currency().Grapheme

			return fmt.Sprintf("%0.3f", amount)
		},
		"GetExtPrice": func(part core.UberPart) string {
			pb := part.GetUnitPrice(part.Quantity)
			amount := model.GetAPIFloatFromString(pb.Price) * model.APIFloat(part.Quantity)
			return fmt.Sprintf("%0.3f", amount)
		},
		"GetStock": func(part core.UberPart) model.APIUint {
			return part.GetAvailabilityAsNumber()
		},
		"InStock": func(part core.UberPart) bool {
			return part.InStock()
		},
		"GetTotal": func(parts []core.UberPart) string {
			if len(parts) == 0 {
				return "N/A"
			}

			total := model.APIFloat(0)
			grapheme := money.New(0, parts[0].GetUnitPrice(1).Currency).Currency().Grapheme
			for _, v := range parts {
				total += v.GetUnitPriceAsNumber(v.Quantity) * model.APIFloat(v.Quantity)
			}

			return fmt.Sprintf("%0.2f %s", total, grapheme)
		},
	}

	tmpl, err := template.New("output").Funcs(funcMap).Parse(`
{{- $datasheetImg := "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAABgAAAAYCAYAAADgdz34AAAABHNCSVQICAgIfAhkiAAAAAlwSFlzAAAAsQAAALEBxi1JjQAAABl0RVh0U29mdHdhcmUAd3d3Lmlua3NjYXBlLm9yZ5vuPBoAAAElSURBVEiJ7daxSgNBEMbxX2IqwQjRUixEbAOKpSj4AIpg4zNIXsDatILa2toExBeIaGUhorYWFqm1MIKVhUX2xByXy90lZf4w3LAs3+x8CztHMoshiuRDqeIpxELOvBoXqyQUqOE55Ms58xq6/8VKMfFjbOE7pcM0pnGLo2gh6mAWc9jGYfgWoY1zXOADnyW9y7nGA1Zwhy+85BSvYwabeMU6dpPuAB5xgiZaGQv86NnbRxkd7CQINXGfUTxOK2h2Kvotim8qyr5gUXkEkUykWTQKmSwahYlFQ5lYNJQ/iwa9RWuSZ0Ua9aTFaB5UMY9LNLCXUzziCqc4wDu60Sm7Idp6r2jRgbMRNN6ihanYhpsQq6HgGZZy5I3QRSpjHfqDGNtvyy+f+FuyXkNG9gAAAABJRU5ErkJggg==" }}
{{- $mouserImg := "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAABgAAAAYCAYAAADgdz34AAAABHNCSVQICAgIfAhkiAAAAAlwSFlzAAAAsQAAALEBxi1JjQAAABl0RVh0U29mdHdhcmUAd3d3Lmlua3NjYXBlLm9yZ5vuPBoAAAFMSURBVEiJ7dY7LwVRFAXgb7xqURGPQiESvURCPCJoJGqNilqiERpKuksiIlHo/AOdV6MU0aAU/AkFhXOZzJ2ZO0NuZyWTOXvtdfbaZ86ZyURq0YWhFL4MXvAAUSIxhh1c/9FgGFNJchFn6MiYFGEW0ymNJXFRHbTEyGXM4D1j0i5a0Ra626hjUmMgpziMYNRX9zdFiqcZ5OEKB2F83giDTUyiqVEGxDavKJpSuE6sly1UxuAQq5grUWcOlzjNE1WX346FEsXzatXswUrG+NeIP6JtDGA+xGuJ+3zIF9FsZRlOoIJ+P8u8CHEl5MexV0eTiT68hQmVwFVC/BbyvXj19S7sZ2i+kdyDZzziJMbdhmsp5OEpaKIcDdKPaVHU+6L+2QA+GmkQ/RsUMfgVjnDv589iKMRHMc0x7jCYo8lEM3oSXE/gq2hBdx0N+ATkJz96/KY9lAAAAABJRU5ErkJggg==" }}
{{- $imageImg := "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAABgAAAAYCAYAAADgdz34AAAABHNCSVQICAgIfAhkiAAAAAlwSFlzAAAAvQAAAL0BHVrG+gAAABl0RVh0U29mdHdhcmUAd3d3Lmlua3NjYXBlLm9yZ5vuPBoAAACaSURBVEiJ7dW9CsIADATgrx1d1VU3fyaH4jv5SkWfSdBFXV3VSdGxDo3gYp0qVHoQCBeOy2UJTUeCG85Rg+CP6EXt0cEI25jPgn9gUqHvQoF5DBZRgltGn2L9ttg6OFhV6ItUzUhwxUUZcahM9IrYx055ojE2oZvhgDumOH3QdyFDXsPyObLaT9QatAatwT8YtP/gK37yD5qNJ58rQBUzA3bAAAAAAElFTkSuQmCC"}}
{{- $linkAnchors := "target=\"_blank\" rel=\"noopener noreferrer\"" }}
{{- $sortIcon := "&nbsp;&nbsp;&#8693;" }}
<!DOCTYPE html>
<html lang="en-US">
	<head>
		<title>{{ .Title }}</title>
		<script>
			function sortTable(n) {
				var table, rows, switching, i, x, y, shouldSwitch, dir, switchcount = 0;
				table = document.getElementById("partsTable");
				switching = true;
				// Set the sorting direction to ascending:
				dir = "asc";
				/* Make a loop that will continue until
				no switching has been done: */
				while (switching) {
					// Start by saying: no switching is done:
					switching = false;
					rows = table.rows;
					/* Loop through all table rows (except the
					first, which contains table headers): */
					for (i = 1; i < (rows.length - 1); i++) {
						// Start by saying there should be no switching:
						shouldSwitch = false;
						/* Get the two elements you want to compare,
						one from current row and one from the next: */
						x = rows[i].getElementsByTagName("TD")[n];
						y = rows[i + 1].getElementsByTagName("TD")[n];
						/* Check if the two rows should switch place,
						based on the direction, asc or desc: */
						if (dir == "asc") {
							if (isNaN(x.innerHTML) || isNaN(y.innerHTML)) {
								if (x.innerHTML.toLowerCase() > y.innerHTML.toLowerCase()) {
									// If so, mark as a switch and break the loop:
									shouldSwitch = true;
									break;
								}
							} else {
								if (Number(x.innerHTML) > Number(y.innerHTML)) {
									//if so, mark as a switch and break the loop:
									shouldSwitch = true;
									break;
								}
							}
						} else if (dir == "desc") {
							if (isNaN(x.innerHTML) || isNaN(y.innerHTML)) {
								if (x.innerHTML.toLowerCase() < y.innerHTML.toLowerCase()) {
									// If so, mark as a switch and break the loop:
									shouldSwitch = true;
									break;
								}
							} else {
								if (Number(x.innerHTML) < Number(y.innerHTML)) {
									//if so, mark as a switch and break the loop:
									shouldSwitch = true;
									break;
								}
							}
						}
					}
					if (shouldSwitch) {
						/* If a switch has been marked, make the switch
						and mark that a switch has been done: */
						rows[i].parentNode.insertBefore(rows[i + 1], rows[i]);
						switching = true;
						// Each time a switch is done, increase this count by 1:
						switchcount++;
					} else {
						/* If no switching has been done AND the direction is "asc",
						set the direction to "desc" and run the while loop again. */
						if (switchcount == 0 && dir == "asc") {
							dir = "desc";
							switching = true;
						}
					}
				}
			}
		</script>
		<style>
			table{
				border-collapse: collapse;
				font-size: 0.8em;
			}
			table, td, th{
				border: 1px solid black;
			}
			th {
				background-color: #0039b3;
				color: white;
			}
			th.sortable {
				cursor: pointer;
			}
			tr {
				color: #000000;
			}
			tr:nth-child(even) {
  				background-color: #e6eeff;
			}
			tr:nth-child(odd) {
  				background-color: #ccdcff;
			}
			tr:hover {
				//background-color: #802000;
				background-color: #efefef;
				color: #000000;
			}
			table {
				width: 100%;
			}
			td.center {
				text-align: center;
			}
			td.bold {
				font-weight: bold;
			}
			td.total {
				padding: 0.2em 0.4em;
				font-weight: bold;
				font-size: 1.2em;
				background-color: #efefef;
			}
			td.small {
				font-size: 0.85em;
			}
			td.ok {
				color: #00802b;
			}
			td.warning {
				color: #ff7733;
			}
			td.error {
				color: #e60000;
			}
			tr.error {
				background-color: #ffb3b3;
			}
			img {
				width: 17px;
			}
			#total {
				text-align: right;
				margin: -1px 0 0 0;
				padding: 0;
				border: 1px solid black;
				background-color: #0039b3;
				color: white;
				font-size: 0.8em;
			}
			#total span {
				margin: 0;
				padding: 0;
					display: inline-block;
				font-weight: bold;
			}
			#total span.totalTitle {
				border-right: 1px solid black;
				padding-right: 5px;
			}
			#total span.totalAmount {
				text-align: right;
				min-width: 100px;
			}
			.tooltip {
				display:inline-block;
				position:relative;
				text-align:left;
			}
			.tooltip h3 {margin:12px 0;}
			.tooltip .left {
				min-width:50px;
				max-width:100px;
				top:50%;
				left:100%;
				margin-right:0;
				margin-left:0;
				transform:translate(-120%, 0);
				padding:0;
				color:#EEEEEE;
				background-color:#ffffff;
				font-weight:normal;
				font-size:13px;
				border-radius:8px;
				position:absolute;
				z-index:99999999;
				box-sizing:border-box;
				box-shadow:0 1px 8px rgba(0,0,0,0.5);
				visibility:hidden; opacity:0; transition:opacity 0.8s;
			}
			.tooltip:hover .left {
				visibility:visible; opacity:1;
			}
			.tooltip .left img {
				width:100px;
				border-radius:8px 8px 0 0;
			}
			.tooltip .text-content {
				padding:10px 20px;
			}
			.tooltip .left i {
				position:absolute;
				top:50%;
				right:100%;
				margin-top:-12px;
				width:12px;
				height:24px;
				overflow:hidden;
			}
			.tooltip .left i::after {
				content:'';
				position:absolute;
				width:12px;
				height:12px;
				left:0;
				top:50%;
				transform:translate(50%,-50%) rotate(-45deg);
				background-color:#444444;
				box-shadow:0 1px 8px rgba(0,0,0,0.5);
			}
			div#thanks {
				font-size: 0.8em;
				text-align: right;
				margin-top: 0.4em;
			}
		</style>
	</head>
	<body>
	<h1>{{ .Title }}</h1>
	<table id="partsTable">
		<tr>
			<th class="sortable" onclick="sortTable(0)">Qty{{ $sortIcon }}</th>
			<th class="sortable" onclick="sortTable(1)">Schematic name{{ $sortIcon }}</th>
			<th class="sortable" onclick="sortTable(2)">Device{{ $sortIcon }}</th>
			<th class="sortable" onclick="sortTable(3)">Value{{ $sortIcon }}</th>
			<th class="sortable" onclick="sortTable(4)">Description{{ $sortIcon }}</th>
			<th class="sortable" onclick="sortTable(5)">Mouser Ref.{{ $sortIcon }}</th>
			<th class="sortable" onclick="sortTable(6)">Stock{{ $sortIcon }}</th>
			<th class="sortable" onclick="sortTable(7)">Unit price{{ $sortIcon }}</th>
			<th class="sortable" onclick="sortTable(8)">Ext. price{{ $sortIcon }}</th>
			<th>Datasheet</th>
			<th>Details</th>
			<th>Image</th>
		</tr>
		{{- range $k, $v := .Parts }}
		{{- $stock := GetStock . }}
		{{- $inStock := InStock . }}
		<tr {{ if not $inStock }}class="error"{{ end }}>
			<td class="center bold">{{ .Quantity }}</td>
			<td>{{ .Parts }}</td>
			<td>{{ .Device }}</td>
			<td>{{ .Value }}</td>
			<td class="small">{{ .Part.Description }}</td>
			<td class="small">{{ .MouserRef }}</td>
			<td class="center {{ if or (lt $stock .Quantity) (not $inStock) }}error{{ else }}ok{{ end }}">{{ $stock }}</td>
			<td class="center">{{ GetPrice . }}</td>
			<td class="center">{{ GetExtPrice . }}</td>
			<td class="center"><a {{ $linkAnchors }} href="{{ .DatasheetURL }}"><img alt="datasheet" width="20" src="{{ $datasheetImg }}" /></a></td>
			<td class="center"><a {{ $linkAnchors }} href="{{ .ProductDetailURL }}"><img alt="details" width="20" src="{{ $mouserImg }}" /></a></td>
			<td class="center">
			{{- if .ImagePath }}
				<div class="tooltip">
					<img alt="image" width="20" src="{{ $imageImg }}" />
					<div class="left">
						<img alt="product image" src="{{ .ImagePath }}" />
					</div>
				</div>
			{{- end }}
			</td>
		</tr>
		{{- end }}
	</table>
	<div id="total">
			<span class="totalTitle">Total</span><span class="totalAmount">{{ GetTotal .Parts }}</span>
	</div>
	<div id="thanks">Icons made by <a {{ $linkAnchors }} href="https://www.flaticon.com/authors/payungkead" title="Payungkead">Payungkead</a> from <a {{ $linkAnchors }} href="https://www.flaticon.com/" title="Flaticon">www.flaticon.com</a></div>
	</body>
</html>`)
	if err != nil {
		return err
	}

	err = tmpl.ExecuteTemplate(w, "output", *o)
	if err != nil {
		return err
	}

	return nil
}
