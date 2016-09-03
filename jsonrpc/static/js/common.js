/**
 * Created by sammy on 16-8-16.
 */

$(document).ready(function(){

    var $tab_tx  = $("#tab_tx"),
        $tab_bal = $("#tab_bal");

    var $div_trans    =  $("div.trans"),
        $div_balances = $("div.balances");

    var selectedClassName = "selected";

    var submitForm = function(){
        var $form = $("form");
        var data = getFormData($form);
        var count = data.count;

        if (!count) {
            count = 1
        }

        for(var i = 1;i <= count;i++) {

            $.ajax({
                contentType: "application/json",
                type: "POST",
                dataType: "json",
                url: "/trans",
                data: JSON.stringify(data),
                success: function (result) {
                    console.log(result);
                    if (result.Code == 1) {
                        //alert("提交成功!");
                        $(".status").html("提交成功")
                    } else {
                        $(".status").html("提交失败，余额不足");
                    }

                    // location.reload();
                },
                error: function (err) {
                    if (err) {
                        console.log(err);
                        $("input[type=reset]").trigger("click");
                    }
                    return false;
                }
            });
        }
    }
    
    
    $("input[type=button]").click(submitForm);

    $tab_tx.click(function(){
        $div_trans.show();
        $div_balances.hide();
        $(this).addClass(selectedClassName).siblings().removeClass(selectedClassName)

    });
    $tab_bal.click(function(){
        $div_trans.hide();
        $div_balances.show();
        $(this).addClass(selectedClassName).siblings().removeClass(selectedClassName)
    });
});

function getFormData($form){
    var unindexed_array = $form.serializeArray();
    var indexed_array = {};

    $.map(unindexed_array, function(n, i){
        indexed_array[n['name']] = n['value'];
    });

    return indexed_array;
}
