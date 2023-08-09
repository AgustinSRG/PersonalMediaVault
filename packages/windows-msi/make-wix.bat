
call heat dir "PersonalMediaVault" -gg -cg PComponentGroup -dr ProgramFiles64Folder -out HeatFile.wxs

call candle -ext WixUIExtension -ext WixUtilExtension -arch x64 HeatFile.wxs -o HeatFile.wixobj

call candle -ext WixUIExtension -ext WixUtilExtension -arch x64 Product.wxs -o Product.wixobj

call light Product.wixobj -spdb HeatFile.wixobj -cultures:en-us -loc en-us.wxl -b PersonalMediaVault -o PersonalMediaVault-1.8.7-x64.msi -ext WixUIExtension -ext WixUtilExtension
call light Product.wixobj -spdb HeatFile.wixobj -cultures:es-es -loc es-es.wxl -b PersonalMediaVault -o PersonalMediaVault-1.8.7-x64-es.msi -ext WixUIExtension -ext WixUtilExtension
