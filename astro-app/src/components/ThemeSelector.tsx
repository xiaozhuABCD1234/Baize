import { useState } from "react";
import {
	Select,
	SelectContent,
	SelectGroup,
	SelectItem,
	SelectTrigger,
	SelectValue,
} from "@/components/ui/select";

export function ThemeSelector() {
	const [theme, setTheme] = useState<string>("system");

	return (
		<Select value={theme} onValueChange={setTheme}>
			<SelectTrigger className="w-45">
				<SelectValue placeholder="Theme" />
			</SelectTrigger>
			<SelectContent>
				<SelectGroup>
					<SelectItem value="light">Light</SelectItem>
					<SelectItem value="dark">Dark</SelectItem>
					<SelectItem value="system">System</SelectItem>
				</SelectGroup>
			</SelectContent>
		</Select>
	);
}
