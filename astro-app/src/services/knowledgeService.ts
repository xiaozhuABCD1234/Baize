export interface KnowledgeItem {
  id: string;
  title: string;
  description: string;
  icon: string;
  details: {
    origin: string;
    history: string;
    features: string;
    significance: string;
  };
  inheritors?: {
    id: string;
    name: string;
    title: string;
    bio: string;
    contributions: string;
  }[];
}

// 模拟从API获取数据
export async function fetchKnowledgeItems(): Promise<KnowledgeItem[]> {
  try {
    // 这里可以替换为实际的API调用
    // const response = await fetch('https://api.example.com/knowledge');
    // const data = await response.json();
    
    // 上海非遗数据
    const mockData: KnowledgeItem[] = [
      {
        id: "1",
        title: "上海剪纸",
        description: "上海剪纸是一种传统民间艺术，以其精细的工艺和独特的风格著称，历史悠久，是上海地区重要的非物质文化遗产。",
        icon: "<svg xmlns=\"http://www.w3.org/2000/svg\" class=\"w-12 h-12 text-muted-foreground\" fill=\"none\" viewBox=\"0 0 24 24\" stroke=\"currentColor\"><path stroke-linecap=\"round\" stroke-linejoin=\"round\" stroke-width=\"2\" d=\"M12 6.253v13m0-13C10.832 5.223 9.246 4.5 7.5 4.5S4.168 5.223 3 6.253v13C4.168 18.777 5.754 19.5 7.5 19.5s3.332-.723 4.5-1.747zm0 0C13.168 19.777 14.754 20.5 16.5 20.5s3.332-.723 4.5-1.747v-13C20.832 5.223 19.246 4.5 17.5 4.5s-3.332.723-4.5 1.747z\" /></svg>",
        details: {
          origin: "上海剪纸起源于明代，至今已有400多年历史。",
          history: "上海剪纸最初是民间艺人用于装饰和祭祀的工具，后来逐渐发展成为一种独立的艺术形式。",
          features: "上海剪纸以精细、秀丽、典雅著称，题材广泛，包括人物、花鸟、山水等。",
          significance: "上海剪纸是上海地区民俗文化的重要组成部分，也是中国传统民间艺术的瑰宝。"
        },
        inheritors: [
          {
            id: "p1",
            name: "陈少芳",
            title: "国家级非物质文化遗产代表性传承人",
            bio: "陈少芳，1945年生，上海剪纸代表性传承人，从事剪纸艺术60余年。",
            contributions: "她创新了多种剪纸技法，作品多次在国内外展览中获奖，为上海剪纸的传承和发展做出了重要贡献。"
          }
        ]
      },
      {
        id: "2",
        title: "上海豫园灯会",
        description: "豫园灯会是上海传统的民俗活动，每年元宵节期间举办，以其精美的灯彩和浓郁的节日氛围闻名，是上海的文化名片。",
        icon: "<svg xmlns=\"http://www.w3.org/2000/svg\" class=\"w-12 h-12 text-muted-foreground\" fill=\"none\" viewBox=\"0 0 24 24\" stroke=\"currentColor\"><path stroke-linecap=\"round\" stroke-linejoin=\"round\" stroke-width=\"2\" d=\"M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10\" /></svg>",
        details: {
          origin: "豫园灯会起源于明代嘉靖年间，至今已有400多年历史。",
          history: "最初是为了庆祝元宵节，后来逐渐发展成为上海地区最具影响力的民俗活动之一。",
          features: "豫园灯会以其精美的灯彩、丰富的题材和浓郁的节日氛围著称，每年吸引大量游客。",
          significance: "豫园灯会是上海地区民俗文化的重要组成部分，也是中国传统节日文化的重要载体。"
        },
        inheritors: [
          {
            id: "p2",
            name: "张灯",
            title: "上海市非物质文化遗产代表性传承人",
            bio: "张灯，1950年生，豫园灯会制作技艺代表性传承人，从事灯彩制作50余年。",
            contributions: "他创新了多种灯彩制作技法，设计制作了许多经典灯组，为豫园灯会的传承和发展做出了重要贡献。"
          }
        ]
      },
      {
        id: "3",
        title: "上海评弹",
        description: "上海评弹是一种传统说唱艺术，融合了说、噱、弹、唱等多种表演形式，具有独特的艺术魅力，是江南地区的代表性曲艺形式。",
        icon: "<svg xmlns=\"http://www.w3.org/2000/svg\" class=\"w-12 h-12 text-muted-foreground\" fill=\"none\" viewBox=\"0 0 24 24\" stroke=\"currentColor\"><path stroke-linecap=\"round\" stroke-linejoin=\"round\" stroke-width=\"2\" d=\"M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z\" /></svg>",
        details: {
          origin: "上海评弹起源于清代乾隆年间，至今已有200多年历史。",
          history: "它是在苏州评弹的基础上发展起来的，逐渐形成了自己独特的风格。",
          features: "上海评弹以说、噱、弹、唱为主要表演形式，语言生动，音乐优美。",
          significance: "上海评弹是江南地区曲艺文化的重要组成部分，也是中国传统说唱艺术的瑰宝。"
        },
        inheritors: [
          {
            id: "p3",
            name: "余红仙",
            title: "国家级非物质文化遗产代表性传承人",
            bio: "余红仙，1939年生，上海评弹代表性传承人，从事评弹艺术70余年。",
            contributions: "她的演唱风格独特，表演功底深厚，为上海评弹的传承和发展做出了重要贡献。"
          },
          {
            id: "p4",
            name: "秦建国",
            title: "上海市非物质文化遗产代表性传承人",
            bio: "秦建国，1956年生，上海评弹代表性传承人，从事评弹艺术50余年。",
            contributions: "他的表演风格严谨，技艺精湛，为上海评弹的传承和发展做出了重要贡献。"
          }
        ]
      },
      {
        id: "4",
        title: "上海龙华庙会",
        description: "龙华庙会是上海历史悠久的传统民俗活动，每年农历三月初三举行，集宗教、文化、商业于一体，是上海地区重要的文化遗产。",
        icon: "<svg xmlns=\"http://www.w3.org/2000/svg\" class=\"w-12 h-12 text-muted-foreground\" fill=\"none\" viewBox=\"0 0 24 24\" stroke=\"currentColor\"><path stroke-linecap=\"round\" stroke-linejoin=\"round\" stroke-width=\"2\" d=\"M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15\" /></svg>",
        details: {
          origin: "龙华庙会起源于明代万历年间，至今已有400多年历史。",
          history: "最初是为了纪念龙华寺的开山祖师，后来逐渐发展成为集宗教、文化、商业于一体的民俗活动。",
          features: "龙华庙会以其丰富的民俗活动、特色小吃和传统手工艺品著称，每年吸引大量游客。",
          significance: "龙华庙会是上海地区民俗文化的重要组成部分，也是中国传统庙会文化的重要代表。"
        },
        inheritors: [
          {
            id: "p5",
            name: "王阿婆",
            title: "上海市非物质文化遗产代表性传承人",
            bio: "王阿婆，1940年生，龙华庙会传统手工艺代表性传承人，从事传统手工艺品制作60余年。",
            contributions: "她制作的传统手工艺品深受游客喜爱，为龙华庙会的传承和发展做出了重要贡献。"
          }
        ]
      },
      {
        id: "5",
        title: "上海老饭店本帮菜制作技艺",
        description: "上海老饭店本帮菜是上海地区的传统菜系，以浓油赤酱、咸淡适中、保持原味为特点，是上海饮食文化的重要组成部分。",
        icon: "<svg xmlns=\"http://www.w3.org/2000/svg\" class=\"w-12 h-12 text-muted-foreground\" fill=\"none\" viewBox=\"0 0 24 24\" stroke=\"currentColor\"><path stroke-linecap=\"round\" stroke-linejoin=\"round\" stroke-width=\"2\" d=\"M12 14l9-5-9-5-9 5 9 5z\" /><path stroke-linecap=\"round\" stroke-linejoin=\"round\" stroke-width=\"2\" d=\"M12 14l6.16-3.422a12.022 12.022 0 01.665 6.479A11.959 11.959 0 0112 20.055a11.96 11.96 0 01-6.824-2.998 12.022 12.022 0 01.665-6.479L12 14z\" /></svg>",
        details: {
          origin: "上海老饭店本帮菜起源于清代道光年间，至今已有180多年历史。",
          history: "它是在江南传统菜系的基础上发展起来的，逐渐形成了自己独特的风格。",
          features: "上海老饭店本帮菜以浓油赤酱、咸淡适中、保持原味为特点，注重原料的新鲜和口感的层次。",
          significance: "上海老饭店本帮菜是上海饮食文化的重要组成部分，也是中国传统菜系的瑰宝。"
        },
        inheritors: [
          {
            id: "p6",
            name: "李伯荣",
            title: "国家级非物质文化遗产代表性传承人",
            bio: "李伯荣，1932年生，上海老饭店本帮菜制作技艺代表性传承人，从事烹饪艺术70余年。",
            contributions: "他创新了多种本帮菜制作技法，培养了许多优秀的厨师，为上海老饭店本帮菜的传承和发展做出了重要贡献。"
          }
        ]
      },
      {
        id: "6",
        title: "上海南翔小笼制作技艺",
        description: "南翔小笼是上海的传统名点，以皮薄馅多、汤汁鲜美著称，制作工艺精细，是上海饮食文化的代表之一。",
        icon: "<svg xmlns=\"http://www.w3.org/2000/svg\" class=\"w-12 h-12 text-muted-foreground\" fill=\"none\" viewBox=\"0 0 24 24\" stroke=\"currentColor\"><path stroke-linecap=\"round\" stroke-linejoin=\"round\" stroke-width=\"2\" d=\"M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4\" /></svg>",
        details: {
          origin: "南翔小笼起源于清代光绪年间，至今已有130多年历史。",
          history: "它最初是南翔镇的特色小吃，后来逐渐发展成为上海的代表性名点。",
          features: "南翔小笼以皮薄馅多、汤汁鲜美著称，制作工艺精细，需要经过多道工序。",
          significance: "南翔小笼是上海饮食文化的重要组成部分，也是中国传统小吃的瑰宝。"
        },
        inheritors: [
          {
            id: "p7",
            name: "黄明贤",
            title: "国家级非物质文化遗产代表性传承人",
            bio: "黄明贤，1943年生，南翔小笼制作技艺代表性传承人，从事小笼制作60余年。",
            contributions: "他创新了多种小笼制作技法，培养了许多优秀的小笼师傅，为南翔小笼的传承和发展做出了重要贡献。"
          }
        ]
      },
      {
        id: "7",
        title: "上海顾绣",
        description: "顾绣是上海地区的传统刺绣工艺，起源于明代，以针法精细、色彩艳丽、题材广泛著称，是中国四大名绣之一。",
        icon: "<svg xmlns=\"http://www.w3.org/2000/svg\" class=\"w-12 h-12 text-muted-foreground\" fill=\"none\" viewBox=\"0 0 24 24\" stroke=\"currentColor\"><path stroke-linecap=\"round\" stroke-linejoin=\"round\" stroke-width=\"2\" d=\"M12 6V4m0 2a2 2 0 100 4m0-4a2 2 0 110 4m-6 8a2 2 0 100-4m0 4a2 2 0 110-4m0 4v2m0-6V4m6 6v10m6-2a2 2 0 100-4m0 4a2 2 0 110-4m0 4v2m0-6V4\" /></svg>",
        details: {
          origin: "顾绣起源于明代嘉靖年间，至今已有400多年历史。",
          history: "它是由上海顾名世家族的女眷创立的，后来逐渐发展成为一种独立的刺绣艺术形式。",
          features: "顾绣以针法精细、色彩艳丽、题材广泛著称，注重表现对象的神韵和细节。",
          significance: "顾绣是上海地区传统工艺的重要组成部分，也是中国传统刺绣艺术的瑰宝。"
        },
        inheritors: [
          {
            id: "p8",
            name: "戴明教",
            title: "国家级非物质文化遗产代表性传承人",
            bio: "戴明教，1932年生，顾绣代表性传承人，从事刺绣艺术70余年。",
            contributions: "她创新了多种顾绣技法，作品多次在国内外展览中获奖，为顾绣的传承和发展做出了重要贡献。"
          }
        ]
      },
      {
        id: "8",
        title: "上海京剧",
        description: "京剧是中国的国粹，上海是京剧的重要发展地之一，上海京剧以其独特的表演风格和艺术特色，成为上海文化的重要组成部分。",
        icon: "<svg xmlns=\"http://www.w3.org/2000/svg\" class=\"w-12 h-12 text-muted-foreground\" fill=\"none\" viewBox=\"0 0 24 24\" stroke=\"currentColor\"><path stroke-linecap=\"round\" stroke-linejoin=\"round\" stroke-width=\"2\" d=\"M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z\" /></svg>",
        details: {
          origin: "上海京剧起源于清代同治年间，至今已有150多年历史。",
          history: "它是在徽班进京的基础上发展起来的，逐渐形成了自己独特的风格。",
          features: "上海京剧以其独特的表演风格、精美的舞台设计和丰富的剧目著称。",
          significance: "上海京剧是上海文化的重要组成部分，也是中国传统戏曲艺术的瑰宝。"
        },
        inheritors: [
          {
            id: "p9",
            name: "尚长荣",
            title: "国家级非物质文化遗产代表性传承人",
            bio: "尚长荣，1940年生，上海京剧代表性传承人，从事京剧艺术70余年。",
            contributions: "他的表演风格独特，技艺精湛，为上海京剧的传承和发展做出了重要贡献。"
          },
          {
            id: "p10",
            name: "李炳淑",
            title: "国家级非物质文化遗产代表性传承人",
            bio: "李炳淑，1942年生，上海京剧代表性传承人，从事京剧艺术70余年。",
            contributions: "她的演唱风格优美，表演功底深厚，为上海京剧的传承和发展做出了重要贡献。"
          }
        ]
      },
      {
        id: "9",
        title: "上海道教音乐",
        description: "上海道教音乐是上海地区的传统音乐形式，具有浓郁的地方特色，是道教文化的重要组成部分，也是上海非物质文化遗产的重要内容。",
        icon: "<svg xmlns=\"http://www.w3.org/2000/svg\" class=\"w-12 h-12 text-muted-foreground\" fill=\"none\" viewBox=\"0 0 24 24\" stroke=\"currentColor\"><path stroke-linecap=\"round\" stroke-linejoin=\"round\" stroke-width=\"2\" d=\"M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z\" /></svg>",
        details: {
          origin: "上海道教音乐起源于唐代，至今已有1000多年历史。",
          history: "它是在传统道教音乐的基础上发展起来的，逐渐形成了自己独特的风格。",
          features: "上海道教音乐以其独特的旋律、丰富的乐器和庄严的仪式感著称。",
          significance: "上海道教音乐是道教文化的重要组成部分，也是中国传统音乐的瑰宝。"
        },
        inheritors: [
          {
            id: "p11",
            name: "陈莲笙",
            title: "上海市非物质文化遗产代表性传承人",
            bio: "陈莲笙，1917年生，上海道教音乐代表性传承人，从事道教音乐80余年。",
            contributions: "他整理和传承了大量道教音乐曲目，为上海道教音乐的传承和发展做出了重要贡献。"
          }
        ]
      },
      {
        id: "10",
        title: "上海民间故事",
        description: "上海民间故事是上海地区劳动人民创作的口头文学作品，反映了上海地区的历史、文化和社会生活，是上海非物质文化遗产的重要组成部分。",
        icon: "<svg xmlns=\"http://www.w3.org/2000/svg\" class=\"w-12 h-12 text-muted-foreground\" fill=\"none\" viewBox=\"0 0 24 24\" stroke=\"currentColor\"><path stroke-linecap=\"round\" stroke-linejoin=\"round\" stroke-width=\"2\" d=\"M12 6.253v13m0-13C10.832 5.223 9.246 4.5 7.5 4.5S4.168 5.223 3 6.253v13C4.168 18.777 5.754 19.5 7.5 19.5s3.332-.723 4.5-1.747zm0 0C13.168 19.777 14.754 20.5 16.5 20.5s3.332-.723 4.5-1.747v-13C20.832 5.223 19.246 4.5 17.5 4.5s-3.332.723-4.5 1.747z\" /></svg>",
        details: {
          origin: "上海民间故事的起源可以追溯到古代，是上海地区劳动人民长期创作和传承的结果。",
          history: "它反映了上海地区的历史、文化和社会生活，是上海地区民俗文化的重要组成部分。",
          features: "上海民间故事以其生动的情节、丰富的想象力和浓郁的地方特色著称。",
          significance: "上海民间故事是上海地区非物质文化遗产的重要组成部分，也是中国民间文学的瑰宝。"
        },
        inheritors: [
          {
            id: "p12",
            name: "顾希佳",
            title: "上海市非物质文化遗产代表性传承人",
            bio: "顾希佳，1941年生，上海民间故事代表性传承人，从事民间文学研究和整理60余年。",
            contributions: "他整理和出版了大量上海民间故事，为上海民间故事的传承和发展做出了重要贡献。"
          }
        ]
      },
      {
        id: "11",
        title: "上海传统建筑营造技艺",
        description: "上海传统建筑营造技艺是上海地区的传统建筑工艺，包括木作、砖作、石作等多种技艺，是上海传统建筑文化的重要组成部分。",
        icon: "<svg xmlns=\"http://www.w3.org/2000/svg\" class=\"w-12 h-12 text-muted-foreground\" fill=\"none\" viewBox=\"0 0 24 24\" stroke=\"currentColor\"><path stroke-linecap=\"round\" stroke-linejoin=\"round\" stroke-width=\"2\" d=\"M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z\" /></svg>",
        details: {
          origin: "上海传统建筑营造技艺的起源可以追溯到古代，是上海地区劳动人民长期实践和积累的结果。",
          history: "它融合了江南传统建筑技艺和西方建筑元素，逐渐形成了自己独特的风格。",
          features: "上海传统建筑营造技艺包括木作、砖作、石作等多种技艺，注重工艺的精细和美观。",
          significance: "上海传统建筑营造技艺是上海传统建筑文化的重要组成部分，也是中国传统建筑工艺的瑰宝。"
        },
        inheritors: [
          {
            id: "p13",
            name: "张宝庆",
            title: "上海市非物质文化遗产代表性传承人",
            bio: "张宝庆，1945年生，上海传统建筑营造技艺代表性传承人，从事建筑营造60余年。",
            contributions: "他精通多种传统建筑营造技艺，参与了许多古建筑的修复和保护工作，为上海传统建筑营造技艺的传承和发展做出了重要贡献。"
          }
        ]
      },
      {
        id: "12",
        title: "上海传统中医药",
        description: "上海传统中医药是上海地区的传统医药文化，包括中医理论、中药炮制、针灸推拿等多种技艺，是上海非物质文化遗产的重要组成部分。",
        icon: "<svg xmlns=\"http://www.w3.org/2000/svg\" class=\"w-12 h-12 text-muted-foreground\" fill=\"none\" viewBox=\"0 0 24 24\" stroke=\"currentColor\"><path stroke-linecap=\"round\" stroke-linejoin=\"round\" stroke-width=\"2\" d=\"M19.428 15.428a2 2 0 00-1.022-.547l-2.387-.477a6 6 0 00-3.86.517l-.318.158a6 6 0 01-3.86.517L6.05 15.21a2 2 0 00-1.806.547M8 4h8l-1 1v5.172a2 2 0 00.586 1.414l5 5c1.26 1.26.367 3.414-1.415 3.414H4.828c-1.782 0-2.674-2.154-1.414-3.414l5-5A2 2 0 009 10.172V5L8 4z\" /></svg>",
        details: {
          origin: "上海传统中医药的起源可以追溯到古代，是上海地区劳动人民长期实践和积累的结果。",
          history: "它融合了传统中医药理论和地方特色，逐渐形成了自己独特的风格。",
          features: "上海传统中医药包括中医理论、中药炮制、针灸推拿等多种技艺，注重整体观念和辨证施治。",
          significance: "上海传统中医药是上海地区非物质文化遗产的重要组成部分，也是中国传统医药文化的瑰宝。"
        },
        inheritors: [
          {
            id: "p14",
            name: "裘沛然",
            title: "国家级非物质文化遗产代表性传承人",
            bio: "裘沛然，1916年生，上海传统中医药代表性传承人，从事中医药临床和研究80余年。",
            contributions: "他在中医药理论和临床实践方面取得了丰硕成果，为上海传统中医药的传承和发展做出了重要贡献。"
          },
          {
            id: "p15",
            name: "张镜人",
            title: "国家级非物质文化遗产代表性传承人",
            bio: "张镜人，1923年生，上海传统中医药代表性传承人，从事中医药临床和研究70余年。",
            contributions: "他在中医药理论和临床实践方面取得了丰硕成果，为上海传统中医药的传承和发展做出了重要贡献。"
          }
        ]
      }
    ];
    
    return mockData;
  } catch (error) {
    console.error('Error fetching knowledge items:', error);
    return [];
  }
}

// 获取单个科普详情
export async function fetchKnowledgeItem(id: string): Promise<KnowledgeItem | null> {
  try {
    // 这里可以替换为实际的API调用
    // const response = await fetch(`https://api.example.com/knowledge/${id}`);
    // const data = await response.json();
    
    // 暂时使用模拟数据，后续可以替换为实际API
    const items = await fetchKnowledgeItems();
    return items.find(item => item.id === id) || null;
  } catch (error) {
    console.error(`Error fetching knowledge item ${id}:`, error);
    return null;
  }
}
