<?php
$accept = strtolower($_SERVER['HTTP_ACCEPT_LANGUAGE'] ?? 'en');
$lang = explode('-', explode(',', $accept)[0])[0];

if ($lang === 'zh') {
    $file = 'zh.html';
    $contentLanguage = 'zh-CN';
} else {
    $file = 'en.html';
    $contentLanguage = 'en';
}

if (!is_file($file)) {
    $file = is_file('zh.html') ? 'zh.html' : 'en.html';
    $contentLanguage = $file === 'zh.html' ? 'zh-CN' : 'en';
}

header('Content-Type: text/html; charset=utf-8');
header('Content-Language: ' . $contentLanguage);
header('Vary: Accept-Language');

echo file_get_contents($file);
