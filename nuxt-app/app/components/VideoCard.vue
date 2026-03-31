<script setup lang="ts">
interface HeritageItem {
	id: number;
	name: string;
	category: string;
	region: string;
	coverImage: string;
	duration?: string;
}

const props = defineProps<{
	item: HeritageItem;
}>();
</script>

<template>
	<article
		class="relative rounded-2xl overflow-hidden shadow-lg group cursor-pointer"
	>
		<!-- 图片区域 -->
		<div class="relative aspect-4/5 bg-burnt-tea/10">
			<img
				:src="item.coverImage"
				:alt="item.name"
				class="w-full h-full object-cover transition-transform duration-500 group-hover:scale-110"
			/>

			<!-- 播放按钮覆盖层 -->
			<div
				class="absolute inset-0 flex items-center justify-center bg-burnt-tea/30 opacity-0 group-hover:opacity-100 transition-opacity duration-300"
			>
				<div
					class="w-16 h-16 rounded-full bg-silk/90 flex items-center justify-center shadow-xl"
				>
					<UIcon name="i-heroicons-play" class="w-8 h-8 text-vermilion ml-1" />
				</div>
			</div>

			<!-- 视频时长标签 -->
			<div
				v-if="item.duration"
				class="absolute bottom-3 right-3 px-2 py-1 rounded bg-burnt-tea/80 text-silk text-xs font-heiti"
			>
				{{ item.duration }}
			</div>

			<!-- 渐变遮罩 -->
			<div
				class="absolute inset-x-0 bottom-0 h-1/2 bg-linear-to-t from-burnt-tea/80 to-transparent"
			/>
		</div>

		<!-- 信息区域 -->
		<div class="absolute inset-x-0 bottom-0 p-4 text-silk">
			<h3 class="font-kaiti font-bold text-xl mb-1">
				{{ item.name }}
			</h3>
			<div class="flex items-center gap-2 text-sm text-silk/80 font-heiti">
				<span class="px-2 py-0.5 rounded bg-silk/20">
					{{ item.category }}
				</span>
				<span class="flex items-center gap-1">
					<UIcon name="i-heroicons-map-pin" class="w-4 h-4" />
					{{ item.region }}
				</span>
			</div>
		</div>
	</article>
</template>

<style scoped>
.font-kaiti {
	font-family: var(--font-kaiti);
}

.font-heiti {
	font-family: var(--font-heiti);
}
</style>
