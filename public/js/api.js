export async function apiQuery(language, query) {
  const res = await fetch(`/api/query?language=${language}`, {
    method: "POST",
    headers: {
      "Content-Type": "text/plain",
      Accept: "application/json",
    },
    body: query,
  });
  const qres = await res.json();
  const obj = Object.fromEntries(
    qres.values.map((val) => val.map((v, idx) => [qres.columns[idx], v])),
  );
  return obj;
}
