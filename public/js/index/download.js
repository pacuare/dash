import { resultsToTsv } from "./csv.js";

export function downloadResults(results) {
  const csvBlob = new Blob([resultsToTsv(results)], { type: "text/tsv" });

  const url = URL.createObjectURL(csvBlob);

  const link = document.createElement("a");
  link.classList.add("hidden");
  link.href = url;
  link.download = "query.csv";
  document.body.appendChild(link);
  link.click();
  link.remove();
  URL.revokeObjectURL(url);
}
