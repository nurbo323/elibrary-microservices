const API_BASE = import.meta.env.VITE_API_BASE || "http://localhost:8080";

async function request(path, options = {}) {
  const response = await fetch(`${API_BASE}${path}`, {
    headers: {
      "Content-Type": "application/json",
      ...(options.headers || {}),
    },
    ...options,
  });

  if (!response.ok) {
    const text = await response.text();
    throw new Error(text || `Request failed with status ${response.status}`);
  }

  if (response.status === 204) {
    return null;
  }

  return response.json();
}

export const api = {
  createUser: (data) =>
    request("/api/users", {
      method: "POST",
      body: JSON.stringify(data),
    }),

  listUsers: () => request("/api/users"),

  createBook: (data) =>
    request("/api/books", {
      method: "POST",
      body: JSON.stringify(data),
    }),

  listBooks: () => request("/api/books"),

  listBorrows: () => request("/api/borrows"),

  borrowBook: (data) =>
    request("/api/borrows", {
      method: "POST",
      body: JSON.stringify(data),
    }),

  returnBook: (borrowId) =>
    request(`/api/borrows/${borrowId}/return`, {
      method: "POST",
    }),
};