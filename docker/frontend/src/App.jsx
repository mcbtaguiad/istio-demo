import React from "react";
import { Routes, Route } from "react-router-dom";
import Login from "./pages/Login";
import Welcome from "./pages/Welcome";

export default function App() {
	return (
		<Routes>
			<Route path="/" element={<Login />} />
			<Route path="/welcome" element={<Welcome />} />
		</Routes>
	);
}
