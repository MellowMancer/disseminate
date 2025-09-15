import React from "react";
import { Link } from "react-router-dom";

export default function HomePage() {
  return (
    <div className="min-h-screen flex flex-col items-center justify-center bg-gray-50 p-4 text-center max-w-xl mx-auto">
      <h1 className="text-4xl font-extrabold mb-4 text-gray-900">
        Disseminate
      </h1>
      <p className="text-lg text-gray-700 mb-6">
        Post to multiple social media platforms from a single, easy-to-use dashboard.
        Save time and reach your audience effortlessly.
      </p>
      <nav className="space-x-4">
        <Link
          to="/upload"
          className="px-5 py-3 bg-blue-600 text-white rounded-md hover:bg-blue-700 transition"
        >
          Start Posting
        </Link>
        <Link
          to="/about"
          className="px-5 py-3 text-blue-600 rounded-md border border-blue-600 hover:bg-blue-50 transition"
        >
          About
        </Link>
        <Link
          to="/auth"
          className="px-5 py-3 text-gray-700 rounded-md hover:text-gray-900 transition"
        >
          Login
        </Link>
      </nav>
    </div>
  );
}
