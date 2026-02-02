export function resultsToTsv(arr) {
  if (arr.length == 0) return "";

  const csvEscape = (row) =>
    row
      .map((v) => v?.toString() ?? "")
      .map((v) => v.replace('"', '""'))
      .map((v) =>
        ["\t", "\n", "\r"].some((c) => v.includes(c)) ? `"${v}"` : v,
      );

  return (
    csvEscape(arr.columns).join("\t") +
    "\r\n" +
    arr.values.map((v) => csvEscape(v).join("\t")).join("\r\n")
  );
}
