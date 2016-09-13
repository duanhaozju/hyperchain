/**
 * Created by sammy on 16-9-13.
 */

function SummaryService($resource) {
    console.log("=========SummaryService==========");
    return {
        getLastestBlock: function(){
            $resource("http://localhost:8084",{},{
                getBlock:{
                    method:"POST"
                }
            }).getBlock({
                method: "block_lastestBlock",
                id:1
            },function(res){
                console.log(res);
            })
        }
    }
}

angular
    .module('starter')
    .factory('SummaryService', SummaryService);