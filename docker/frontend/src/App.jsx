import React from "react";
import { BrowserRouter, Routes, Route } from "react-router-dom";
import Login from "./pages/Login";
import Welcome from "./pages/Welcome";

export default function App() {
	return (
		// Set basename to /app
		<BrowserRouter basename="/app">
			<Routes>
				<Route path="/" element={<Login />} />          {/* /app */}
				<Route path="/welcome" element={<Welcome />} /> {/* /app/welcome */}
			</Routes>
		</BrowserRouter>
	);
}
