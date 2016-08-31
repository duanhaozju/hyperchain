/**
 * Created by sammy on 16-8-16.
 */

function getFormData($form){
    var unindexed_array = $form.serializeArray();
    var indexed_array = {};

    $.map(unindexed_array, function(n, i){
        indexed_array[n['name']] = n['value'];
    });

    console.log(indexed_array)
    return indexed_array;
}


$(document).ready(function(){
    $("input[type=button]").click(function(){
        // var url = $("input[name='address']").val();
        var $form = $("form");

        // if(!url){
        //     url = location.host
        // }
        //for(var i=1;i<500;i++){

        $.ajax({
            // contentType: "application/json; charset=utf-8",
            type:"POST",
            dataType: "json",
            url: "/trans",
            data: getFormData($form),
            success: function( result ) {
                console.log(result);
                if(result.Code == 1){
                    //alert("提交成功!");
                    $(".status").html("提交成功")
                } else {
                    $(".status").html("提交成功")
//                    alert("交易验证失败，您没有足够的金额！");
                }

//                    $("input[type=reset]").trigger("click");
               // location.reload();
            },
            error: function(err){
                if(err){
                    console.log(err);
                    $("input[type=reset]").trigger("click");
                }
                return false;
            }
        });
    });
});
