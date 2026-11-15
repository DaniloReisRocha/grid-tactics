export function backendURL(path: string) {
  const apiURL = import.meta.env.VITE_API_URL as string | undefined;
  const url = apiURL ? new URL(apiURL, window.location.href)
    : new URL(`${window.location.protocol === "https:" ? "https:" : "http:"}//${window.location.hostname}:8080`);
  url.pathname = path; url.search = ""; url.hash = "";
  return url;
}

export function backendHeaders(url: URL, initial?: HeadersInit) {
  const headers = new Headers(initial);
  if (!headers.has("Accept")) headers.set("Accept", "application/json");
  if ([".ngrok-free.app", ".ngrok.app", ".ngrok.io"].some((suffix) => url.hostname.endsWith(suffix))) {
    headers.set("ngrok-skip-browser-warning", "1");
  }
  return headers;
}
