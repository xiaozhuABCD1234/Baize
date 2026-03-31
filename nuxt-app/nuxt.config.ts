// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
	compatibilityDate: "2025-07-15",
	devtools: { enabled: false },
	modules: ["@nuxt/ui"],
	css: ["~/assets/css/main.css"],
	app: {
		head: {
			title: "非遗百科 - 传承文化之美",
			meta: [
				{
					name: "description",
					content: "非遗百科首页，探索中国传统非物质文化遗产",
				},
			],
		},
	},
});
