import React from "react"; // ADD THIS
import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";

const API = "/api";

export default function Welcome() {
	const [username, setUsername] = useState("");
	const [group, setGroup] = useState("");
	const [version, setVersion] = useState("");

	const navigate = useNavigate();

	useEffect(() => {
		const token = localStorage.getItem("token");
		if (!token) {
			navigate("/");
			return;
		}

		// Load version
		fetch(API + "/version")
			.then(res => res.json())
			.then(data => setVersion(data.version));

		// Load profile
		fetch(API + "/profile", { headers: { Authorization: "Bearer " + token } })
			.then(res => {
				if (!res.ok) throw new Error();
				return res.json();
			})
			.then(data => {
				setUsername(data.username);
				setGroup(data.group);
			})
			.catch(() => {
				localStorage.removeItem("token");
				navigate("/");
			});
	}, []);

	function logout() {
		localStorage.removeItem("token");
		navigate("/");
	}

	return (
		<div>
			<h2>Istio Demo</h2>

			<p><b>Backend Version:</b> {version}</p>
			<p><b>Welcome, {username}</b></p>
			<p><b>Your group:</b> {group}</p>

			<button onClick={logout}>Logout</button>
		</div>
	);
}
