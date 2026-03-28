import { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";

const API = "/api"; // points to your Go backend via reverse proxy

export default function Login() {
	const [user, setUser] = useState("");
	const [pass, setPass] = useState("");
	const [msg, setMsg] = useState("");
	const [version, setVersion] = useState("");

	const navigate = useNavigate();

	useEffect(() => {
		fetch(API + "/version")
			.then(res => res.json())
			.then(data => setVersion(data.version));
	}, []);

	async function login() {
		const res = await fetch(API + "/login", {
			method: "POST",
			headers: { "Content-Type": "application/json" },
			body: JSON.stringify({ username: user, password: pass })
		});

		if (!res.ok) {
			setMsg("Login failed");
			return;
		}

		const data = await res.json();
		localStorage.setItem("token", data.token);
		navigate("/welcome");
	}

	async function register() {
		const res = await fetch(API + "/register", {
			method: "POST",
			headers: { "Content-Type": "application/json" },
			body: JSON.stringify({ username: user, password: pass })
		});

		const text = await res.text();
		setMsg(res.ok ? "Registered: " + text : "Error: " + text);
	}

	return (
		<div>
			<h2>Istio Demo</h2>
			<p><b>Backend Version:</b> {version}</p>

			<input placeholder="username" onChange={e => setUser(e.target.value)} />
			<input type="password" placeholder="password" onChange={e => setPass(e.target.value)} />

			<br /><br />

			<button onClick={login}>Login</button>
			<button onClick={register}>Register</button>

			<pre>{msg}</pre>
		</div>
	);
}
