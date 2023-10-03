<template>
    <div id="app">
        <Bookmarks />
    </div>
</template>

<script charset="utf-8">

import Bookmarks from './components/Bookmarks.vue'

export default {
    name: 'app',
    components: {
        Bookmarks
    },
    data() {
        return {
            msg: 'Hello World Gosuki'
        }
    }
}
</script>
