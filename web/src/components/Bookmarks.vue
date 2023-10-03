<template>
  <div class="container">
    <div class="search-wrapper">
      <input type="search" v-model="search" placeholder="Search bookmarks..."/>
    </div>
    <!--<div v-for="bookmark in bookmarkList" v-bind:key="bookmarkList">
      <p>{{bookmark}}</p>
    </div>-->
    <hr/>
    <div class="wrapper">
      <div class="card" v-for="bookmark in filteredList"  v-bind:key="filterBookmark">
        <a v-bind:href="bookmark.url" target="_blank">
          <div class="url">{{ bookmark.url }}</div>
          <div class="tags" v-if="bookmark.tags.toString()"><small> Tags : {{ bookmark.tags.toString() }}</small></div>
          <div class="metadata"> {{ bookmark.metadata }} </div>
        </a>
      </div>
    </div>
  </div>
</template>




<script>
import axios from 'axios'
import {Bookmark} from './Bookmark.js'

  export default {
    name: 'Bookmarks',
    search: '',
    data () {
      return {
        bookmarkList: [],
        search: '',
      }
    },

    // Fetches posts when the component is created.
    mounted() {
    axios
      .get('http://localhost:4242/api/urls')
      .then(response => (
        console.log('TEST ', response.data.bookmarks),
        this.bookmarkList = response.data.bookmarks.map( a => new Bookmark(a)),
        console.log('TEST GET API', this.bookmarkList)
        ))
      .catch(error => console.log('GET Error: ', error))
    },

    computed: {
      filteredList() {
        return this.bookmarkList.filter(filterBookmark => {
          console.log('this.bookmarks',this.bookmarks)
          console.log('filterbookmark', typeof(filterBookmark.tags.toString()))
          let filterMetadata = filterBookmark.metadata.toLowerCase()
          let filterTags = filterBookmark.tags.toString().toLowerCase()
          let filtredMetadata = filterMetadata.includes(this.search.toLowerCase())
          let filterdTags = filterTags.includes(this.search.toLowerCase())
          return  filtredMetadata || filterdTags
        })
      }
    }
  }
</script>



 
<style>
h3 {
  margin-bottom: 5%;
  }
.search-wrapper {
   position: relative;
   padding-bottom: 2%;
}
 .search-wrapper input {
	 padding: 4px 12px;
	 color: rgba(0, 0, 0, .70);
	 border: 1px solid rgba(0, 0, 0, .12);
	 transition: 0.15s all ease-in-out;
	 background: white;
}
 .search-wrapper input:focus {
	 outline: none;
	 transform: scale(1.05);
}
 .search-wrapper input:focus + label {
	 font-size: 10px;
	 transform: translateY(-24px) translateX(-12px);
}
 .search-wrapper input::-webkit-input-placeholder {
	 font-size: 12px;
	 color: rgba(0, 0, 0, .50);
	 font-weight: 100;
}
 .wrapper {
	 display: flex;
	 max-width: 100%;
	 flex-wrap: wrap;
	 padding-top: 2%;
     flex-direction: column;
}
 .card {
	/* box-shadow: rgba(0, 0, 0, 0.117647) 0px 1px 6px, rgba(0, 0, 0, 0.117647) 0px 1px 4px;*/
	margin: 1%;
    padding: 0.3rem;
	transition: 100ms all ease-in-out;
}
 .card:hover {
	 transform: translateX(+10px);
}
 .card a {
	 display: flex;
	 flex-direction: column;
   text-decoration: none;
   height: 100%;
	 color: #03a9f4;
	 
}
 .card a small {
	 font-size: 10px;
	 padding: 4px;
}

.card .metadata {
  /*height: 50%;*/
}
 
</style>

