/**
 * Created by sammy on 16-8-16.
 */

var selectedClassName = "selected";

$(document).ready(function(){

    var $tab_tx    = $("#tab_tx"),
        $tab_bal   = $("#tab_bal"),
        $tab_block = $("#tab_block");

    var $div_trans    =  $("div.trans"),
        $div_balances = $("div.balances"),
        $div_blocks   = $("div.blocks");

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
    };
    
    $("input[type=button]").click(submitForm);

    $tab_tx.click({div: $div_trans, url: "/trans"},tabClickHandler);
    $tab_bal.click({div: $div_balances, url: "/balances"},tabClickHandler);
    $tab_block.click({div: $div_blocks, url: "/blocks"},tabClickHandler);

});


function getFormData($form){
    var unindexed_array = $form.serializeArray();
    var indexed_array = {};

    $.map(unindexed_array, function(n, i){
        indexed_array[n['name']] = n['value'];
    });

    return indexed_array;
}

// Tab click event Handler
function tabClickHandler(event) {

    // console.log(event.data);
    var _this = $(this);

    var $div = $(event.data.div[0]);

    $.ajax({
        contentType: "application/json",
        type: "GET",
        dataType: "json",
        url: event.data.url,
        success: function (result) {
            // console.log(result);

            if (result.Code == 0) {
                alert("error");
                return
            }
            var $tbody = $div.find("tbody");
            $tbody.html("");

            if (result.Data instanceof Array) {
                // Array
                for (var i = 0; i < result.Data.length; i++) {
                    var $tr = $("<tr>");

                    for (var col in result.Data[i]) {
                        var $td = $("<td>");

                        $td.html(result.Data[i][col]);

                        $tr.append($td);
                    }

                    $tbody.append($tr);
                }
            } else {
                // Object
                for (var key in result.Data) {
                    var $tr = $("<tr>");
                    var $td1 = $("<td>");
                    var $td2 = $("<td>");

                    $td1.html(key);
                    $td2.html(result.Data[key]);

                    $tr.append($td1);
                    $tr.append($td2);

                    $tbody.append($tr);
                }
            }

            $div.show().siblings().hide();
            _this.addClass(selectedClassName).siblings().removeClass(selectedClassName)
        },
        error: function (err) {
            if (err) {
                console.log(err);
            }
            return false;
        }
    });

}


