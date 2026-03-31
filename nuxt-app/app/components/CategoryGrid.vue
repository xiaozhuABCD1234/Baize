<script setup lang="ts">
interface Category {
	id: string;
	title: string;
	icon: string;
	color: string;
	bgColor: string;
	description: string;
}

const categories: Category[] = [
	{
		id: "encyclopedia",
		title: "非遗百科",
		icon: "i-heroicons-book-open",
		color: "text-icon-cyan",
		bgColor: "bg-icon-cyan/10",
		description: "了解非遗知识",
	},
	{
		id: "identify",
		title: "辨别真伪",
		icon: "i-heroicons-shield-check",
		color: "text-icon-purple",
		bgColor: "bg-icon-purple/10",
		description: "鉴别非遗真伪",
	},
	{
		id: "gallery",
		title: "作品广场",
		icon: "i-heroicons-photo",
		color: "text-icon-green",
		bgColor: "bg-icon-green/10",
		description: "欣赏非遗作品",
	},
	{
		id: "map",
		title: "地图找非遗",
		icon: "i-heroicons-map-pin",
		color: "text-icon-orange",
		bgColor: "bg-icon-orange/10",
		description: "探索附近非遗",
	},
];
</script>

<template>
	<section class="py-8">
		<h2 class="text-2xl font-kaiti font-bold text-burnt-tea mb-6 text-center">
			发现非遗
		</h2>

		<div class="grid grid-cols-2 md:grid-cols-4 gap-4 md:gap-6">
			<button
				v-for="category in categories"
				:key="category.id"
				class="group flex flex-col items-center p-6 rounded-2xl transition-all duration-300 hover:scale-105 active:scale-95"
				:class="category.bgColor"
			>
				<!-- 图标容器 - 白描风格 -->
				<div
					class="w-20 h-20 rounded-full flex items-center justify-center mb-4 transition-all duration-300 group-hover:shadow-lg"
					:class="[category.bgColor, category.color]"
				>
					<UIcon :name="category.icon" class="w-10 h-10 ink-style" />
				</div>

				<!-- 标题 -->
				<h3 class="font-kaiti font-bold text-lg text-burnt-tea mb-1">
					{{ category.title }}
				</h3>

				<!-- 描述 -->
				<p class="text-sm text-burnt-tea/70 font-heiti">
					{{ category.description }}
				</p>
			</button>
		</div>
	</section>
</template>

<style scoped>
.font-kaiti {
	font-family: var(--font-kaiti);
}

.font-heiti {
	font-family: var(--font-heiti);
}

.ink-style {
	stroke: currentColor;
	stroke-width: 1.5;
	fill: none;
	stroke-linecap: round;
	stroke-linejoin: round;
}
</style>
