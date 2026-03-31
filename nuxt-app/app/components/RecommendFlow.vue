<script setup lang="ts">
import VideoCard from "./VideoCard.vue";

interface HeritageItem {
	id: number;
	name: string;
	category: string;
	region: string;
	coverImage: string;
	duration?: string;
	rowSpan?: number;
}

// 示例数据 - 使用网络图片
const heritageItems: HeritageItem[] = [
	{
		id: 1,
		name: "剪纸",
		category: "传统美术",
		region: "陕西延安",
		coverImage:
			"https://images.unsplash.com/photo-1604916287784-c324202b3205?w=400&h=500&fit=crop",
		duration: "0:08",
	},
	{
		id: 2,
		name: "皮影戏",
		category: "传统戏剧",
		region: "河北唐山",
		coverImage:
			"https://images.unsplash.com/photo-1596422846543-75c6fc197f07?w=400&h=500&fit=crop",
		duration: "0:12",
	},
	{
		id: 3,
		name: "苏绣",
		category: "传统美术",
		region: "江苏苏州",
		coverImage:
			"https://images.unsplash.com/photo-1583336135672-8e4322eb8089?w=400&h=500&fit=crop",
		duration: "0:10",
	},
	{
		id: 4,
		name: "景泰蓝",
		category: "传统技艺",
		region: "北京",
		coverImage:
			"https://images.unsplash.com/photo-1578662996442-48f60103fc96?w=400&h=500&fit=crop",
		duration: "0:15",
	},
	{
		id: 5,
		name: "川剧变脸",
		category: "传统戏剧",
		region: "四川成都",
		coverImage:
			"https://images.unsplash.com/photo-1537531383496-f4749b8032cf?w=400&h=500&fit=crop",
		duration: "0:09",
	},
	{
		id: 6,
		name: "龙泉青瓷",
		category: "传统技艺",
		region: "浙江龙泉",
		coverImage:
			"https://images.unsplash.com/photo-1565193566173-7a0ee3dbe261?w=400&h=500&fit=crop",
		duration: "0:11",
	},
	{
		id: 7,
		name: "古琴艺术",
		category: "传统音乐",
		region: "江苏扬州",
		coverImage:
			"https://images.unsplash.com/photo-1510915361894-db8b60106cb1?w=400&h=500&fit=crop",
		duration: "0:14",
	},
	{
		id: 8,
		name: "泥人张",
		category: "传统美术",
		region: "天津",
		coverImage:
			"https://images.unsplash.com/photo-1558604668-d06c5de6d73e?w=400&h=500&fit=crop",
		duration: "0:07",
	},
];
</script>

<template>
	<section class="py-8">
		<h2 class="text-2xl font-kaiti font-bold text-burnt-tea mb-6 text-center">
			推荐非遗
		</h2>

		<!-- 瀑布流容器 -->
		<div class="masonry-container">
			<div
				v-for="item in heritageItems"
				:key="item.id"
				class="break-inside-avoid mb-5"
			>
				<VideoCard :item="item" />
			</div>
		</div>

		<!-- 查看更多按钮 -->
		<div class="text-center mt-8">
			<UButton
				size="lg"
				class="font-kaiti bg-vermilion hover:bg-vermilion/90 text-silk font-bold"
				variant="solid"
			>
				查看更多非遗项目
				<template #trailing>
					<UIcon name="i-heroicons-arrow-right" class="w-5 h-5" />
				</template>
			</UButton>
		</div>
	</section>
</template>

<style scoped>
.font-kaiti {
	font-family: var(--font-kaiti);
}
</style>
