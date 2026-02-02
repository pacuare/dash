export async function apiQuery(language, query) {
  const res = await fetch(`/api/query?language=${language}`, {
    method: "POST",
    headers: {
      "Content-Type": "text/plain",
      Accept: "application/json",
    },
    body: query,
  });
  return await res.json();
}
