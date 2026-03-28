const API = "http://localhost:3000/api";

export async function login(username, password) {
	const res = await fetch(API + "/login", {
		method: "POST",
		headers: { "Content-Type": "application/json" },
		body: JSON.stringify({ username, password })
	});

	if (!res.ok) throw new Error("login failed");
	return res.json();
}

export async function register(username, password) {
	return fetch(API + "/register", {
		method: "POST",
		headers: { "Content-Type": "application/json" },
		body: JSON.stringify({ username, password })
	});
}

export async function getProfile(token) {
	const res = await fetch(API + "/profile", {
		headers: { Authorization: "Bearer " + token }
	});

	if (!res.ok) throw new Error("unauthorized");
	return res.json();
}
