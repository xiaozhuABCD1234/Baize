<script setup lang="ts">
const searchQuery = ref("");
const isListening = ref(false);

const handleMicClick = () => {
	isListening.value = !isListening.value;
};
</script>

<template>
	<div class="w-full py-6">
		<div class="flex items-center gap-4 w-full">
			<!-- 搜索框区域 -->
			<div class="flex-1 max-w-[90%] relative">
				<UInput
					v-model="searchQuery"
					size="xl"
					class="search-input"
					:ui="{
						base: 'w-full text-[18px] font-heiti text-burnt-tea',
					}"
					placeholder="搜索非遗项目、传承人、地区..."
				>
					<template #leading>
						<UIcon
							name="i-heroicons-magnifying-glass"
							class="w-6 h-6 text-burnt-tea/60"
						/>
					</template>
				</UInput>
			</div>

			<!-- 麦克风按钮 -->
			<button
				@click="handleMicClick"
				class="shrink-0 w-16 h-16 rounded-full flex items-center justify-center transition-all duration-300"
				:class="[
					isListening
						? 'bg-vermilion text-silk mic-pulse'
						: 'bg-vermilion/90 text-silk hover:bg-vermilion',
				]"
				title="语音搜索"
			>
				<UIcon
					:name="isListening ? 'i-heroicons-stop' : 'i-heroicons-microphone'"
					class="w-8 h-8"
				/>
			</button>
		</div>

		<!-- 语音搜索提示 -->
		<Transition
			enter-active-class="transition-all duration-300 ease-out"
			enter-from-class="opacity-0 -translate-y-2"
			enter-to-class="opacity-100 translate-y-0"
			leave-active-class="transition-all duration-200 ease-in"
			leave-from-class="opacity-100 translate-y-0"
			leave-to-class="opacity-0 -translate-y-2"
		>
			<div
				v-if="isListening"
				class="mt-3 text-center text-vermilion font-kaiti text-lg"
			>
				请说话，松开结束录音...
			</div>
		</Transition>
	</div>
</template>

<style scoped>
.font-heiti {
	font-family: var(--font-heiti);
}

.font-kaiti {
	font-family: var(--font-kaiti);
}
</style>
