/**
 * Created by sammy on 16-8-16.
 */

function getFormData($form){
    var unindexed_array = $form.serializeArray();
    var indexed_array = {};

    $.map(unindexed_array, function(n, i){
        indexed_array[n['name']] = n['value'];
    });

    return indexed_array;
}


$(document).ready(function(){
    $("input[type=button]").click(function(){
        // var url = $("input[name='address']").val();
        var $form = $("form");

        // if(!url){
        //     url = location.host
        // }

        $.ajax({
            type:"POST",
            dataType: "json",
            url: "/trans",
            data: getFormData($form),
            success: function( result ) {
                console.log(result);
                if(result.Code == 1){
                    alert("提交成功!");
                } else {
                    alert("交易验证失败，您或许没有足够的金额！");
                }

//                    $("input[type=reset]").trigger("click");
                location.reload();
            },
            error: function(err){
                if(err){
                    console.log(err);
                    alert("the "+url+" dont open server");
                    $("input[type=reset]").trigger("click");
                }
                return false;
            }
        });
    });
});
