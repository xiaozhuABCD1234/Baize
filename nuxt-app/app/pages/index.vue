<script setup lang="ts">
useHead({
	title: "非遗百科 - 传承文化之美",
});
</script>

<template>
	<div class="min-h-screen bg-silk">
		<!-- 顶部留白区域 -->
		<div class="pt-8" />

		<!-- 搜索与语音区域 -->
		<div class="px-4 md:px-8">
			<UContainer>
				<SearchHeader />
			</UContainer>
		</div>

		<!-- 核心金刚区 -->
		<UContainer>
			<CategoryGrid />
		</UContainer>

		<!-- 分隔线 -->
		<div class="px-4 md:px-8 my-4">
			<UContainer>
				<div class="border-t-2 border-burnt-tea/20 my-4" />
			</UContainer>
		</div>

		<!-- 推荐流 -->
		<UContainer class="pb-12">
			<RecommendFlow />
		</UContainer>

		<!-- 页脚提示 -->
		<div class="text-center pb-8 text-burnt-tea/60 text-sm font-heiti">
			<p>长按识别二维码 · 关注非遗百科</p>
		</div>
	</div>
</template>

<style scoped>
.font-heiti {
	font-family: var(--font-heiti);
}
</style>
