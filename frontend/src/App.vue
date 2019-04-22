<template>
  <div id="app">
    <div class="navigation">
      <div v-for="file in files" class="tab" @click="changeFile(file)">
        {{file}}
      </div>
    </div>
    <div class="chart-container">
      <div class="header">{{currentOpenFile}}</div>
      <line-chart :data="chartData"></line-chart>
    </div>
  </div>
</template>


<script>
  export default {
    data() {
      return {
        files: [],
        currentOpenFile: "",
        chartData: []
      };
    },
    created() {
      this.$socket.onopen = () => {
        this.$options.sockets.onmessage = (msg) => {
          console.log(event);
          this.handleEvent(msg.data);
        }
      }
    },
    methods: {
      handleEvent(data) {
        try {
          const event = JSON.parse(data);

          switch (event.type) {
            case 'root_subscribe':
              {
                this.files = event.data;
                break;
              }
            case 'file_subscribe':
              {
                this.chartData = event.data.values;
                break;
              }
          }
        } catch (err) {
          console.error(err)
        }
      },
      changeFile(fileName) {
        this.currentOpenFile = fileName;

        this.$socket.sendObj({
          type: "file_subscribe",
          data: {
            "name": fileName
          }
        })
      },
    }
  }
</script>

<style>
  body {
    margin: 0;
  }

  #app {
    font-family: Helvetica, sans-serif;
    text-align: center;
    display: flex;
    height: 100vh;
    width: 100%;
    overflow: hidden;
  }

  .navigation {
    width: 30%;
    border-right: 2px solid black;
  }

  .navigation .tab {
    margin: 20px;
    padding: 20px;
    border: 1px solid black;
    cursor: pointer;
  }

  .chart-container {
    width: 70%;
    display: flex;
    flex-direction: column;
  }

  .chart-container .header {
    height: 30%;
    margin: 30px;
  }
</style>