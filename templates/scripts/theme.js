(function () {

    const btnTheme = document.getElementById('theme');

    const btnLanguage = document.getElementById('language-switch');
    const currentLanguage = localStorage.getItem('language');

    // 获取浏览器首选语言
    const userLang = navigator.language || navigator.userLanguage;
    // 获取用户所有首选语言，按优先级排序
    const userLangs = navigator.languages || [userLang];
    const prefersLang = userLangs[0]

    // console.log("prefersLang:" + prefersLang);

    // Check for dark mode preference at the OS level
    const prefersDarkScheme = window.matchMedia('(prefers-color-scheme: dark)');

    // Get the user's theme preference from local storage, if it's available
    const currentTheme = localStorage.getItem('theme');
    // If the user's preference in localStorage is dark...
    const themeColorMetaTag = document.getElementById('theme-color');
    if (currentTheme == 'dark') {
        // ...let's toggle the .dark-theme class on the body
        document.body.classList.toggle('dark-theme');
        themeColorMetaTag.setAttribute('content', '#121212')
        // Otherwise, if the user's preference in localStorage is light...
    } else if (currentTheme == 'light') {
        // ...let's toggle the .light-theme class on the body
        document.body.classList.toggle('light-theme');
        themeColorMetaTag.setAttribute('content', '#fff')
    }

    if (currentLanguage == 'zh') {
        switchLanguageOnload('zh');
        language = 'zh';
    } else if (currentLanguage == 'en') {
        switchLanguageOnload('en');
        language = 'en';
    } else if (prefersLang == 'zh-CN') {
        switchLanguageOnload('zh');
        language = 'zh';
    } else {
        switchLanguageOnload('en');
        language = 'en';
    }

    // Listen for a click on the button
    btnTheme.addEventListener('click', function () {
        // If the user's OS setting is dark and matches our .dark-theme class...
        if (prefersDarkScheme.matches) {
            // ...then toggle the light mode class
            document.body.classList.toggle('light-theme');
            // ...but use .dark-theme if the .light-theme class is already on the body,
            var theme = document.body.classList.contains('light-theme') ? 'light' : 'dark';
        } else {
            // Otherwise, let's do the same thing, but for .dark-theme
            document.body.classList.toggle('dark-theme');
            var theme = document.body.classList.contains('dark-theme') ? 'dark' : 'light';
        }
        const themeColorMetaTag = document.getElementById('theme-color');
        if (theme == 'light') {
            themeColorMetaTag.setAttribute('content', '#fff')
        } else {
            themeColorMetaTag.setAttribute('content', '#121212')
        }
        // Finally, let's save the current preference to localStorage to keep using it
        localStorage.setItem('theme', theme);
    });

    btnLanguage.addEventListener('click', function () {
        if (language == 'zh') {
            switchLanguage('en');
            language = 'en';
        } else {
            switchLanguage('zh');
            language = 'zh';
        }
        localStorage.setItem('language', language);
    });
})();