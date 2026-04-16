import { useState } from "react";
import { authService } from "@/services/authService";
import type { AuthUser } from "@/services/authService";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";

export function LoginForm() {
	const [email, setEmail] = useState("");
	const [password, setPassword] = useState("");
	const [error, setError] = useState("");
	const [loading, setLoading] = useState(false);

	const handleSubmit = async (e: React.FormEvent) => {
		e.preventDefault();
		setError("");

		if (!email || !password) {
			setError("请填写所有字段");
			return;
		}

		setLoading(true);
		try {
			const data = await authService.login(email, password);
			authService.setToken(data.token);
			authService.setUser(data as AuthUser);
			window.location.href = "/";
		} catch (err: any) {
			setError(err.message || "登录失败");
		} finally {
			setLoading(false);
		}
	};

	return (
		<form onSubmit={handleSubmit} className="space-y-5">
			{error && (
				<div className="p-3 rounded-lg bg-red-50 border border-red-200 text-red-600 text-sm">
					{error}
				</div>
			)}

			<div className="space-y-1.5">
				<label className="text-sm font-medium text-foreground">邮箱</label>
				<div className="relative">
					<div className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground">
						<svg
							xmlns="http://www.w3.org/2000/svg"
							width="18"
							height="18"
							viewBox="0 0 24 24"
							fill="none"
							stroke="currentColor"
							strokeWidth="2"
							strokeLinecap="round"
							strokeLinejoin="round"
						>
							<path d="M4 4h16c1.1 0 2 .9 2 2v12c0 1.1-.9 2-2 2H4c-1.1 0-2-.9-2-2V6c0-1.1.9-2 2-2z"></path>
							<polyline points="22,6 12,13 2,6"></polyline>
						</svg>
					</div>
					<Input
						type="email"
						placeholder="请输入您的邮箱"
						value={email}
						onChange={(e) => setEmail(e.target.value)}
						className="pl-10"
					/>
				</div>
			</div>

			<div className="space-y-1.5">
				<label className="text-sm font-medium text-foreground">密码</label>
				<div className="relative">
					<div className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground">
						<svg
							xmlns="http://www.w3.org/2000/svg"
							width="18"
							height="18"
							viewBox="0 0 24 24"
							fill="none"
							stroke="currentColor"
							strokeWidth="2"
							strokeLinecap="round"
							strokeLinejoin="round"
						>
							<rect x="3" y="11" width="18" height="11" rx="2" ry="2"></rect>
							<path d="M7 11V7a5 5 0 0 1 10 0v4"></path>
						</svg>
					</div>
					<Input
						type="password"
						placeholder="请输入密码"
						value={password}
						onChange={(e) => setPassword(e.target.value)}
						className="pl-10"
					/>
				</div>
			</div>

			<div className="flex items-center justify-between text-sm">
				<label className="flex items-center gap-2 cursor-pointer">
					<input
						type="checkbox"
						className="w-4 h-4 rounded text-primary focus:ring-primary/30 border-border"
					/>
					<span className="text-muted-foreground">记住我</span>
				</label>
				<a
					href="/forget"
					className="text-primary hover:underline transition-all"
				>
					忘记密码？
				</a>
			</div>

			<Button
				type="submit"
				disabled={loading}
				className="w-full py-2.5 bg-primary text-primary-foreground rounded-lg text-sm font-medium hover:opacity-90 transition-all shadow-sm"
			>
				{loading ? "登录中..." : "登录账号"}
			</Button>
		</form>
	);
}
